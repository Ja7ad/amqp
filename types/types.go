package types

import (
	"time"

	rabbit "github.com/rabbitmq/amqp091-go"
)

type (
	ExchangeKind uint8
	DeliverMode  uint8
)

const (
	// Direct exchange delivers messages to queues based on the message routing key. A direct exchange is ideal for the unicast routing of messages. They can be used for multicast routing as well.
	//
	//Here is how it works:
	//
	//    A queue binds to the exchange with a routing key K
	//    When a new message with routing key R arrives at the direct exchange, the exchange routes it to the queue if K = R
	//    If multiple queues are bound to a direct exchange with the same routing key K, the exchange will route the message to all queues for which K = R
	Direct ExchangeKind = iota
	// Fanout exchange routes messages to all of the queues that are bound to it and the routing key is ignored. If N queues are bound to a fanout exchange, when a new message is published to that exchange a copy of the message is delivered to all N queues. Fanout exchanges are ideal for the broadcast routing of messages.
	//
	//Because a fanout exchange delivers a copy of a message to every queue bound to it, its use cases are quite similar:
	//
	//    Massively multi-player online (MMO) games can use it for leaderboard updates or other global events
	//    Sport news sites can use fanout exchanges for distributing score updates to mobile clients in near real-time
	//    Distributed systems can broadcast various state and configuration updates
	//    Group chats can distribute messages between participants using a fanout exchange (although AMQP does not have a built-in concept of presence, so XMPP may be a better choice)
	Fanout
	// Topic exchanges route messages to one or many queues based on matching between a message routing key and the pattern that was used to bind a queue to an exchange. The topic exchange type is often used to implement various publish/subscribe pattern variations. Topic exchanges are commonly used for the multicast routing of messages.
	//
	//Topic exchanges have a very broad set of use cases. Whenever a problem involves multiple consumers/applications that selectively choose which type of messages they want to receive, the use of topic exchanges should be considered.
	//
	//Example uses:
	//
	//    Distributing data relevant to specific geographic location, for example, points of sale
	//    Background task processing done by multiple workers, each capable of handling specific set of tasks
	//    Stocks price updates (and updates on other kinds of financial data)
	//    News updates that involve categorization or tagging (for example, only for a particular sport or team)
	//    Orchestration of services of different kinds in the cloud
	//    Distributed architecture/OS-specific software builds or packaging where each builder can handle only one architecture or OS
	Topic
	// Headers exchange is designed for routing on multiple attributes that are more easily expressed as message headers than a routing key. Headers exchanges ignore the routing key attribute. Instead, the attributes used for routing are taken from the headers attribute. A message is considered matching if the value of the header equals the value specified upon binding.
	//
	//It is possible to bind a queue to a headers exchange using more than one header for matching. In this case, the broker needs one more piece of information from the application developer, namely, should it consider messages with any of the headers matching, or all of them? This is what the "x-match" binding argument is for. When the "x-match" argument is set to "any", just one matching header value is sufficient. Alternatively, setting "x-match" to "all" mandates that all the values must match.
	//
	//For "any" and "all", headers beginning with the string x- will not be used to evaluate matches. Setting "x-match" to "any-with-x" or "all-with-x" will also use headers beginning with the string x- to evaluate matches.
	//
	//Headers exchanges can be looked upon as "direct exchanges on steroids". Because they route based on header values, they can be used as direct exchanges where the routing key does not have to be a string; it could be an integer or a hash (dictionary) for example.
	Headers
)

func (k ExchangeKind) String() string {
	switch k {
	case Direct:
		return "direct"
	case Fanout:
		return "fanout"
	case Topic:
		return "topic"
	case Headers:
		return "headers"
	}
	return ""
}

// Action is an action that occurs after processed this delivery
type Action uint8

type Delivery struct {
	rabbit.Delivery
}

const (
	Transient DeliverMode = iota + 1
	Persistent
)

// ConsumerHandler defines the handler of each Delivery and return Action
//
// vPtr you variable for decode body
type ConsumerHandler func(func(vPtr any) (Delivery, error)) (action Action)

const (
	// Ack default ack this msg after you have successfully processed this delivery.
	Ack Action = iota
	// AckMultiple when multiple is true, this delivery and all prior unacknowledged deliveries on the same channel will be acknowledged.
	// This is useful for batch processing of deliveries.
	AckMultiple
	// NackDiscard the message will be dropped or delivered to a server configured dead-letter queue.
	NackDiscard
	// NackRequeue deliver this message to a different consumer.
	NackRequeue
	// Reject reject message in queue
	Reject
	// RejectRequeue reject message and requeue in queue
	RejectRequeue
)

type Exchange struct {
	Name       string
	Kind       ExchangeKind
	Declare    bool
	Passive    bool
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Arguments  map[string]interface{}
}

type Queue struct {
	Name       string
	Declare    bool
	Passive    bool
	Durable    bool
	Exclusive  bool
	AutoDelete bool
	NoWait     bool
	Arguments  map[string]interface{}
}

type Consumer struct {
	Name      string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Arguments map[string]interface{}
}

type RoutingKey struct {
	Key     string
	Declare bool
}

// Return captures a flattened struct of fields returned by the server when a
// Publishing is unable to be delivered either due to the `mandatory` flag set
// and no route found, or `immediate` flag set and no free consumer.
type Return struct {
	rabbit.Return
}

// Confirmation notifies the acknowledgment or negative acknowledgement of a publishing identified by its delivery tag.
// Use NotifyPublish to consume these events. ReconnectionCount is useful in that each time it increments, the DeliveryTag
// is reset to 0, meaning you can use ReconnectionCount+DeliveryTag to ensure uniqueness
type Confirmation struct {
	rabbit.Confirmation
	ReconnectionCount int
}

type PublisherConfirmation []*rabbit.DeferredConfirmation

type Publishing struct {
	// Application or exchange specific fields,
	// the headers exchange will inspect this field.
	Headers map[string]interface{}

	// Properties
	ContentType     string      // MIME content type
	ContentEncoding string      // MIME content encoding
	DeliveryMode    DeliverMode // Transient (0 or 1) or Persistent (2)
	Priority        uint8       // 0 to 9
	CorrelationId   string      // correlation identifier
	ReplyTo         string      // address to to reply to (ex: RPC)
	Expiration      string      // message expiration spec
	MessageId       string      // message identifier
	Timestamp       time.Time   // message timestamp
	Type            string      // message type name
	UserId          string      // creating user id - ex: "guest"
	AppId           string      // creating application id

	// The application specific payload of the message
	Body any
}

type PublisherConfig struct {
	RetryDelay time.Duration
	MaxRetries int
}

type EncodeType uint8

const (
	JSON = iota
	GOB
	PROTO
)

type Encoder interface {
	Encode(v any) ([]byte, error)
	Decode(data []byte, vPtr any) error
}
