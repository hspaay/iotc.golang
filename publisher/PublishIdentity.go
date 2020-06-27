// Package publisher with publishing of this publisher's identity
package publisher

// PublishIdentity publishes this publisher's identity on startup or update
func (publisher *Publisher) PublishIdentity() {
	identity := publisher.fullIdentity
	publisher.logger.Infof("Publisher.PublishIdentity: publish identity: %s", publisher.fullIdentity.Address)
	publisher.publishObject(identity.Address, true, identity.PublisherIdentityMessage, nil)
}
