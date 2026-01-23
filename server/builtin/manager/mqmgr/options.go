package mqmgr

func WithPublishHeaders(headers map[string]any) PublishOption {
	return func(opts *PublishOptions) {
		opts.Headers = headers
	}
}

func WithPublishDurable(durable bool) PublishOption {
	return func(opts *PublishOptions) {
		opts.Durable = durable
	}
}

func WithSubscribeDurable(durable bool) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.Durable = durable
	}
}

func WithAutoAck(autoAck bool) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.AutoAck = autoAck
	}
}
