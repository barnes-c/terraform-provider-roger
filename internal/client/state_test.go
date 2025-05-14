// SPDX-FileCopyrightText: 2025 CERN
//
// SPDX-License-Identifier: GPL-3.0-or-later

package roger_test

import (
	"testing"

	roger "roger/internal/client"

	"github.com/stretchr/testify/require"
)

func TestStateCRUD(t *testing.T) {
	host := "woger-direct.cern.ch"
	port := 8201

	cli, err := roger.NewClient(host, port)
	require.NoError(t, err)

	hostname := "tf-test-roger-123.cern.ch"

	initialMessage := "Terraform test init"
	initialAppState := "production"

	t.Logf("Creating state for hostname: %s", hostname)
	createdState, err := cli.CreateState(hostname, initialMessage, initialAppState)
	require.NoError(t, err)
	require.Equal(t, hostname, createdState.Hostname)
	require.Equal(t, initialAppState, createdState.AppState)

	t.Log("Reading state...")
	readState, err := cli.GetState(hostname)
	require.NoError(t, err)
	require.Equal(t, createdState.Hostname, readState.Hostname)

	t.Log("Updating state...")
	updatedMessage := "Terraform test updated"
	updatedAppState := "draining"

	updatedState, err := cli.UpdateState(hostname, updatedMessage, updatedAppState)
	require.NoError(t, err)
	require.Equal(t, updatedMessage, updatedState.Message)
	require.Equal(t, updatedAppState, updatedState.AppState)

	t.Log("Final read to confirm update...")
	finalState, err := cli.GetState(hostname)
	require.NoError(t, err)
	require.Equal(t, updatedMessage, finalState.Message)

	t.Log("Deleting state...")
	err = cli.DeleteState(hostname)
	require.NoError(t, err)

	t.Log("Verifying state is deleted...")
	_, err = cli.GetState(hostname)
	require.Error(t, err)
}
