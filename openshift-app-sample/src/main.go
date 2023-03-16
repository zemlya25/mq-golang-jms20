// Copyright (c) IBM Corporation 2021.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0, which is available at
// http://www.eclipse.org/legal/epl-2.0.
//
// SPDX-License-Identifier: EPL-2.0

// Package main provides the entry point for a executable application.
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/zemlya25/mq-golang-jms20/mqjms"
)

func main() {
	fmt.Println("Beginning world!!!")

	portNum, _ := strconv.Atoi(os.Getenv("PORT"))

	// Initialise the attributes of the CF in whatever way you like
	cf := mqjms.ConnectionFactoryImpl{
		QMName:      os.Getenv("QMNAME"),
		Hostname:    os.Getenv("HOSTNAME"),
		PortNumber:  portNum,
		ChannelName: os.Getenv("CHANNELNAME"),
		UserName:    os.Getenv("USERNAME"),
		Password:    os.Getenv("PASSWORD"),
	}

	// Creates a connection to the queue manager, using defer to close it automatically
	// at the end of the function (if it was created successfully)
	context, errCtx := cf.CreateContext()
	if context != nil {
		defer context.Close()
	}

	if errCtx != nil {
		log.Fatal(errCtx)
	} else {
		fmt.Println("  -- Connection successful")

		// TODO - Add your application code here!

	}

	fmt.Println("Ending world!!!")
}
