package storage

import (
	"context"
	"fmt"
	"github.com/Astemirdum/user-app/server/internal/config"
	"github.com/ClickHouse/clickhouse-go/v2"
	"net"
	"time"
)

type Storage struct {
	db clickhouse.Conn
}

func NewStorage(cfg *config.Config) (*Storage, error) {
	clickCfg := cfg.Clickhouse
	opts := &clickhouse.Options{
		Addr: []string{net.JoinHostPort(clickCfg.Host, clickCfg.Port)},
		Auth: clickhouse.Auth{
			Database: clickCfg.NameDB,
			Username: clickCfg.Username,
			Password: clickCfg.Password,
		},
		DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:      time.Duration(10) * time.Second,
		MaxOpenConns:     5,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Duration(10) * time.Minute,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	}
	conn, err := clickhouse.Open(opts)
	if err != nil {
		return nil, err
	}
	if err = conn.Ping(context.Background()); err != nil {
		return nil, err
	}

	st := &Storage{db: conn}

	ctx := context.Background()
	if err = st.Migrate(ctx, cfg.Kafka); err != nil {
		return nil, err
	}

	return st, nil
}

func (st *Storage) Migrate(ctx context.Context, cfg config.Kafka) error {
	if err := st.createKafkaEngine(ctx, cfg); err != nil {
		return err
	}

	return st.createMV(ctx)
}

func (st *Storage) createKafkaEngine(ctx context.Context, cfg config.Kafka) error {
	kafkaEngine := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS user_logs
	(
		raw String
	) ENGINE = Kafka()
	SETTINGS
	kafka_broker_list = '%s',
		kafka_topic_list = '%s',
		kafka_group_name = '%s',
		kafka_format = '%s',
		kafka_num_consumers = %d;`,
		cfg.AddrClick, cfg.Topic, cfg.ConsumerGroup, cfg.Format, cfg.ConsumerNum)

	return st.db.Exec(ctx, kafkaEngine)
}
func (st *Storage) createMV(ctx context.Context) error {
	mv := `CREATE MATERIALIZED VIEW  IF NOT EXISTS user_logs_view
            ENGINE = MergeTree
			ORDER BY tuple()
		AS
		SELECT * FROM user_logs
    		SETTINGS
    		stream_like_engine_allow_direct_select = 1;`

	return st.db.Exec(ctx, mv)
}

func (st *Storage) Close() error {
	return st.db.Close()
}
