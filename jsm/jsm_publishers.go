package jetstream

import (
	"errors"
	"strings"
)

func (s *natsStore) Publish(topic string, message []byte) error {
	if strings.TrimSpace(topic) == empty {
		return errors.New("invalid topic name")
	}
	s.mountAndRegisterPublishTopics(topic)
	_, err := s.jsmClient.Publish(topic, message)
	return err
}

func (s *natsStore) mountAndRegisterPublishTopics(topic string) {
	s.mu.Lock()
	if _, ok := s.publishTopics[topic]; ok {
		s.mu.Unlock()
		return
	}

	if _, ok := s.subscriptions[topic]; ok {
		s.mu.Unlock()
		return
	}

	s.publishTopics[topic] = topic
	s.mu.Unlock()

	s.registerSubjectsOnStream()
}
