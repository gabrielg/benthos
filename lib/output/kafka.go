package output

import (
	"fmt"

	"github.com/Jeffail/benthos/v3/lib/log"
	"github.com/Jeffail/benthos/v3/lib/message/batch"
	"github.com/Jeffail/benthos/v3/lib/metrics"
	"github.com/Jeffail/benthos/v3/lib/output/writer"
	"github.com/Jeffail/benthos/v3/lib/types"
	"github.com/Jeffail/benthos/v3/lib/util/kafka/sasl"
	"github.com/Jeffail/benthos/v3/lib/util/retries"
	"github.com/Jeffail/benthos/v3/lib/util/tls"
	"github.com/Jeffail/benthos/v3/lib/x/docs"
)

//------------------------------------------------------------------------------

func saslFieldSpec() docs.FieldSpec {
	return docs.FieldAdvanced("sasl", "Enables SASL authentication.").WithChildren(
		docs.FieldCommon("enabled", "Whether SASL authentication is enabled."),
		docs.FieldCommon("user", "A plain text username. It is recommended that you use environment variables to populate this field.", "${USER}"),
		docs.FieldCommon("password", "A plain text password. It is recommended that you use environment variables to populate this field.", "${PASSWORD}"),
	)
}

func init() {
	Constructors[TypeKafka] = TypeSpec{
		constructor: NewKafka,
		Description: `
The kafka output type writes a batch of messages to Kafka brokers and waits for
acknowledgement before propagating it back to the input. The config field
` + "`ack_replicas`" + ` determines whether we wait for acknowledgement from all
replicas or just a single broker.

Both the ` + "`key` and `topic`" + ` fields can be dynamically set using
function interpolations described [here](/docs/configuration/interpolation#functions).
When sending batched messages these interpolations are performed per message
part.`,
		sanitiseConfigFunc: func(conf Config) (interface{}, error) {
			return sanitiseWithBatch(conf.Kafka, conf.Kafka.Batching)
		},
		Async:   true,
		Batches: true,
		FieldSpecs: append(docs.FieldSpecs{
			docs.FieldDeprecated("round_robin_partitions"),
			docs.FieldCommon("addresses", "A list of broker addresses to connect to. If an item of the list contains commas it will be expanded into multiple addresses.", []string{"localhost:9092"}, []string{"localhost:9041,localhost:9042"}, []string{"localhost:9041", "localhost:9042"}),
			tls.FieldSpec(),
			sasl.FieldSpec(),
			docs.FieldCommon("topic", "The topic to publish messages to.").SupportsInterpolation(false),
			docs.FieldCommon("client_id", "An identifier for the client connection."),
			docs.FieldCommon("key", "The key to publish messages with.").SupportsInterpolation(false),
			docs.FieldCommon("partitioner", "The partitioning algorithm to use.").HasOptions("fnv1a_hash", "murmur2_hash", "random", "round_robin"),
			docs.FieldCommon("compression", "The compression algorithm to use.").HasOptions("none", "snappy", "lz4", "gzip"),
			docs.FieldCommon("max_in_flight", "The maximum number of parallel message batches to have in flight at any given time."),
			docs.FieldAdvanced("ack_replicas", "Ensure that messages have been copied across all replicas before acknowledging receipt."),
			docs.FieldAdvanced("max_msg_bytes", "The maximum size in bytes of messages sent to the target topic."),
			docs.FieldAdvanced("timeout", "The maximum period of time to wait for message sends before abandoning the request and retrying."),
			docs.FieldAdvanced("target_version", "The version of the Kafka protocol to use."),
			batch.FieldSpec(),
		}, retries.FieldSpecs()...),
	}
}

//------------------------------------------------------------------------------

// NewKafka creates a new Kafka output type.
func NewKafka(conf Config, mgr types.Manager, log log.Modular, stats metrics.Type) (Type, error) {
	k, err := writer.NewKafka(conf.Kafka, mgr, log, stats)
	if err != nil {
		return nil, err
	}
	var w Type
	if conf.Kafka.MaxInFlight == 1 {
		w, err = NewWriter(
			TypeKafka, k, log, stats,
		)
	} else {
		w, err = NewAsyncWriter(
			TypeKafka, conf.Kafka.MaxInFlight, k, log, stats,
		)
	}
	if bconf := conf.Kafka.Batching; err == nil && !bconf.IsNoop() {
		policy, err := batch.NewPolicy(bconf, mgr, log.NewModule(".batching"), metrics.Namespaced(stats, "batching"))
		if err != nil {
			return nil, fmt.Errorf("failed to construct batch policy: %v", err)
		}
		w = NewBatcher(policy, w, log, stats)
	}
	return w, err
}

//------------------------------------------------------------------------------
