/*
 * Copyright (c) IBM Corporation 2019
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0, which is available at
 * http://www.eclipse.org/legal/epl-2.0.
 *
 * SPDX-License-Identifier: EPL-2.0
 */
package main

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zemlya25/mq-golang-jms20/jms20subset"
	"github.com/zemlya25/mq-golang-jms20/mqjms"
)

/*
 * Test receiving a specific message from a queue using its CorrelationID
 */
func TestGetByCorrelID(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	// First, check the queue is empty
	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, conErr := context.CreateConsumer(queue)
	assert.Nil(t, conErr)
	if consumer != nil {
		defer consumer.Close()
	}
	reqMsgTest, err := consumer.ReceiveNoWait()
	assert.Nil(t, err)
	assert.Nil(t, reqMsgTest)

	// Put a couple of messages before the one we're aiming to get back
	context.CreateProducer().SendString(queue, "One")
	context.CreateProducer().SendString(queue, "Two")

	myCorrelID := "MyCorrelID"
	myMsgThreeStr := "Three"
	sentMsg := context.CreateTextMessageWithString(myMsgThreeStr)
	sentMsg.SetJMSCorrelationID(myCorrelID)
	err = context.CreateProducer().Send(queue, sentMsg)
	assert.Nil(t, err)
	sentMsgID := sentMsg.GetJMSMessageID()

	// Put a couple of messages after the one we're aiming to get back
	context.CreateProducer().SendString(queue, "Four")
	context.CreateProducer().SendString(queue, "Five")

	// Create the consumer to read by CorrelID
	correlIDConsumer, correlErr := context.CreateConsumerWithSelector(queue, "JMSCorrelationID = '"+myCorrelID+"'")
	assert.Nil(t, correlErr)
	gotCorrelMsg, correlGetErr := correlIDConsumer.ReceiveNoWait()

	// Clean up the remaining messages from the queue before we start checking if
	// we got the right one back.
	var cleanupMsg jms20subset.Message
	for ok := true; ok; ok = (cleanupMsg != nil) {
		cleanupMsg, err = consumer.ReceiveNoWait()
	}

	// Now do the comparisons
	assert.Nil(t, correlGetErr)
	assert.NotNil(t, gotCorrelMsg)
	gotMsgID := gotCorrelMsg.GetJMSMessageID()
	assert.Equal(t, sentMsgID, gotMsgID)

	switch msg := gotCorrelMsg.(type) {
	case jms20subset.TextMessage:
		assert.Equal(t, myMsgThreeStr, *msg.GetText())
	default:
		fmt.Println(reflect.TypeOf(gotCorrelMsg))
		assert.Fail(t, "Got something other than a text message")
	}

}

/*
 * Test that errors are returned for invalid selector strings.
 */
func TestSelectorParsing(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	queue := context.CreateQueue("DEV.QUEUE.1")

	// Creating a consumer with an empty string selector is equivalent to creating
	// a consumer without a selector - should succeed.
	noSelectorConsumer, noSelectorErr := context.CreateConsumerWithSelector(queue, "")
	assert.Nil(t, noSelectorErr)
	assert.NotNil(t, noSelectorConsumer)
	_, noSelectorErr = noSelectorConsumer.ReceiveNoWait()
	noSelectorConsumer.Close()
	assert.Nil(t, noSelectorErr)

	// Check that we can create a consumer with a CorrelID that matches a messageID
	// which is used in request/reply scenarios.
	correlIDConsumer, correlIDErr := context.CreateConsumerWithSelector(queue, "JMSCorrelationID = '414d5120514d312020202020202020201017155c0255b621'")
	assert.Nil(t, correlIDErr)
	assert.NotNil(t, correlIDConsumer)

	// JMSMessageID selectors are now supported.
	msgIDConsumer, msgIDErr := context.CreateConsumerWithSelector(queue, "JMSMessageID = 'ID:1234'")
	assert.Nil(t, msgIDErr)
	assert.NotNil(t, msgIDConsumer)

	// MessageID selector that has an empty ID
	failMsgIDConsumer, failMsgIDErr := context.CreateConsumerWithSelector(queue, "JMSMessageID = 'ID:'")
	assert.NotNil(t, failMsgIDErr)
	assert.Nil(t, failMsgIDConsumer)

	// Check that we get an appropriate error when trying to create a consumer with
	// a malformed selector.
	fail1Consumer, fail1Err := context.CreateConsumerWithSelector(queue, "JMSCorrelationID")
	assert.NotNil(t, fail1Err)
	assert.Nil(t, fail1Consumer)

	// Check that we get an appropriate error when trying to create a consumer with
	// a malformed selector.
	fail2Consumer, fail2Err := context.CreateConsumerWithSelector(queue, "JMSCorrelationID = ")
	assert.NotNil(t, fail2Err)
	assert.Nil(t, fail2Consumer)

	// Check that we get an appropriate error when trying to create a consumer with
	// a malformed selector.
	fail3Consumer, fail3Err := context.CreateConsumerWithSelector(queue, "JMSCorrelationID = '")
	assert.NotNil(t, fail3Err)
	assert.Nil(t, fail3Consumer)

	// Check that we get an appropriate error when trying to create a consumer with
	// a malformed selector.
	fail4Consumer, fail4Err := context.CreateConsumerWithSelector(queue, "JMSCorrelationID = ''")
	assert.NotNil(t, fail4Err)
	assert.Nil(t, fail4Consumer)

}

/*
 * Test that we can round trip various correlation IDs into and out of the
 * message object successfully.
 */
func TestCorrelIDParsing(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	msg := context.CreateTextMessage()
	assert.Equal(t, "", msg.GetJMSCorrelationID())

	msg.SetJMSCorrelationID("")
	assert.Equal(t, "", msg.GetJMSCorrelationID())

	testCorrel := "Hello World"
	msg.SetJMSCorrelationID(testCorrel)
	assert.Equal(t, testCorrel, msg.GetJMSCorrelationID())

	testCorrel = "  "
	msg.SetJMSCorrelationID(testCorrel)
	assert.Equal(t, testCorrel, msg.GetJMSCorrelationID())

	testCorrel = "010203040506"
	msg.SetJMSCorrelationID(testCorrel)
	assert.Equal(t, testCorrel, msg.GetJMSCorrelationID())

	testCorrel = "ThisIsAVeryLongCorrelationIDWhichIsMoreThanTwentyFourCharacters"
	msg.SetJMSCorrelationID(testCorrel)
	assert.Equal(t, testCorrel[0:12], msg.GetJMSCorrelationID())

	// MessageID format
	testCorrel = "414d5120514d312020202020202020201017155c0255b621"
	msg.SetJMSCorrelationID(testCorrel)
	assert.Equal(t, testCorrel, msg.GetJMSCorrelationID())

	// Empty correlationID
	testCorrel = "000000000000000000000000000000000000000000000000"
	msg.SetJMSCorrelationID(testCorrel)
	assert.Equal(t, "", msg.GetJMSCorrelationID())

}

// Do a round-trip send and receive of a message in order to check the correlation ID
func checkCorrelIDOnSendReceive(t *testing.T, context jms20subset.JMSContext, queue jms20subset.Queue,
	producer jms20subset.JMSProducer, consumer jms20subset.JMSConsumer,
	inputCorrelID string, expectedCorrelID string) {

	msg := context.CreateTextMessage()
	msg.SetJMSCorrelationID(inputCorrelID)
	assert.Equal(t, expectedCorrelID, msg.GetJMSCorrelationID())
	sendErr := producer.Send(queue, msg)
	assert.Nil(t, sendErr)

	rcvMsg, rcvErr := consumer.ReceiveNoWait()
	assert.Nil(t, rcvErr)
	assert.NotNil(t, rcvMsg)
	assert.Equal(t, expectedCorrelID, rcvMsg.GetJMSCorrelationID())

}

/*
 * Test that we can round trip various correlation IDs via messages sent
 * and received, since there is some checking on send.
 */
func TestCorrelIDParsingOnSend(t *testing.T) {

	// Loads CF parameters from connection_info.json and applicationApiKey.json in the Downloads directory
	cf, cfErr := mqjms.CreateConnectionFactoryFromDefaultJSONFiles()
	assert.Nil(t, cfErr)

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, ctxErr := cf.CreateContext()
	assert.Nil(t, ctxErr)
	if context != nil {
		defer context.Close()
	}

	// First, check the Send queue is initially empty
	queue := context.CreateQueue("DEV.QUEUE.1")
	consumer, rConErr := context.CreateConsumer(queue)
	assert.Nil(t, rConErr)
	if consumer != nil {
		defer consumer.Close()
	}
	reqMsgTest, rcvErr := consumer.ReceiveNoWait()
	assert.Nil(t, rcvErr)
	assert.Nil(t, reqMsgTest)

	producer := context.CreateProducer().SetTimeToLive(1000)
	assert.NotNil(t, producer)

	// Try out the various combinations of correlation ID
	testCorrel := ""
	checkCorrelIDOnSendReceive(t, context, queue, producer, consumer, testCorrel, "")

	// MessageID format
	testCorrel = "414d5120514d312020202020202020201017155c0255b621"
	checkCorrelIDOnSendReceive(t, context, queue, producer, consumer, testCorrel, testCorrel)

	// Empty correlationID
	testCorrel = "000000000000000000000000000000000000000000000000"
	checkCorrelIDOnSendReceive(t, context, queue, producer, consumer, testCorrel, "")

	testCorrel = "ThisIsAVeryLongCorrelationIDWhichIsMoreThanTwentyFourCharacters"
	checkCorrelIDOnSendReceive(t, context, queue, producer, consumer, testCorrel, testCorrel[0:12])

	testCorrel = "Hello World"
	checkCorrelIDOnSendReceive(t, context, queue, producer, consumer, testCorrel, testCorrel)

	testCorrel = "  "
	checkCorrelIDOnSendReceive(t, context, queue, producer, consumer, testCorrel, testCorrel)

	testCorrel = "010203040506"
	checkCorrelIDOnSendReceive(t, context, queue, producer, consumer, testCorrel, testCorrel)

}
