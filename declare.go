package amqp

import "github.com/Ja7ad/amqp/types"

func declareQueue(chanManager *channel, queue *types.Queue) error {
	if !queue.Declare {
		return nil
	}

	if queue.Passive {
		_, err := chanManager.queueDeclarePassiveSafe(
			queue.Name,
			queue.Durable,
			queue.AutoDelete,
			queue.Exclusive,
			queue.NoWait,
			queue.Arguments,
		)
		if err != nil {
			return err
		}
		return nil
	}
	_, err := chanManager.queueDeclareSafe(
		queue.Name,
		queue.Durable,
		queue.AutoDelete,
		queue.Exclusive,
		queue.NoWait,
		queue.Arguments,
	)
	if err != nil {
		return err
	}
	return nil
}

func declareExchange(chanManager *channel, exchange *types.Exchange) error {
	if !exchange.Declare {
		return nil
	}

	if exchange.Passive {
		err := chanManager.exchangeDeclarePassiveSafe(
			exchange.Name,
			exchange.Kind.String(),
			exchange.Durable,
			exchange.AutoDelete,
			exchange.Internal,
			exchange.NoWait,
			exchange.Arguments,
		)
		if err != nil {
			return err
		}
		return nil
	}
	err := chanManager.exchangeDeclareSafe(
		exchange.Name,
		exchange.Kind.String(),
		exchange.Durable,
		exchange.AutoDelete,
		exchange.Internal,
		exchange.NoWait,
		exchange.Arguments,
	)
	if err != nil {
		return err
	}
	return nil
}

func declareBindings(chanManager *channel, queueName, exchangeName string, routingKeys []*types.RoutingKey, consumer *types.Consumer) error {
	for _, rk := range routingKeys {
		if !rk.Declare {
			continue
		}

		if err := chanManager.queueBindSafe(
			queueName,
			rk.Key,
			exchangeName,
			consumer.NoWait,
			consumer.Arguments,
		); err != nil {
			return err
		}
	}

	return nil
}
