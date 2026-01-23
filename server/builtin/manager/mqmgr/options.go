package mqmgr

// WithPublishHeaders 设置消息头
func WithPublishHeaders(headers map[string]any) PublishOption {
	return func(opts *PublishOptions) {
		opts.Headers = headers
	}
}

// WithPublishDurable 设置是否持久化
func WithPublishDurable(durable bool) PublishOption {
	return func(opts *PublishOptions) {
		opts.Durable = durable
	}
}

// WithSubscribeDurable 设置队列是否持久化
func WithSubscribeDurable(durable bool) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.Durable = durable
	}
}

// WithAutoAck 设置是否自动确认消息
func WithAutoAck(autoAck bool) SubscribeOption {
	return func(opts *SubscribeOptions) {
		opts.AutoAck = autoAck
	}
}
