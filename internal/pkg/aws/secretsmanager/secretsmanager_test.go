// Copyright 2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package secretsmanager wraps AWS SecretsManager API functionality.
package secretsmanager

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/aws/secretsmanager/mocks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/golang/mock/gomock"

	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/stretchr/testify/require"
)

func TestSecretsManager_CreateSecret(t *testing.T) {
	mockSecretName := "github-token-backend-badgoose"
	mockSecretString := "H0NKH0NKH0NK"
	mockError := errors.New("mockError")
	mockOutput := &secretsmanager.CreateSecretOutput{
		ARN: aws.String("arn-goose"),
	}
	mockAwsErr := awserr.New(secretsmanager.ErrCodeResourceExistsException, "", nil)

	tests := map[string]struct {
		inSecretName   string
		inSecretString string
		callMock       func(m *mocks.MockSecretsManagerAPI)

		expectedError error
	}{
		"should wrap error returned by CreateSecret": {
			inSecretName:   mockSecretName,
			inSecretString: mockSecretString,
			callMock: func(m *mocks.MockSecretsManagerAPI) {
				m.EXPECT().CreateSecret(&secretsmanager.CreateSecretInput{
					Name:         aws.String(mockSecretName),
					SecretString: aws.String(mockSecretString),
				}).Return(nil, mockError)
			},
			expectedError: fmt.Errorf("create secret %s: %w", mockSecretName, mockError),
		},

		"should return no error if secret already exists": {
			inSecretName:   mockSecretName,
			inSecretString: mockSecretString,
			callMock: func(m *mocks.MockSecretsManagerAPI) {
				m.EXPECT().CreateSecret(&secretsmanager.CreateSecretInput{
					Name:         aws.String(mockSecretName),
					SecretString: aws.String(mockSecretString),
				}).Return(nil, mockAwsErr)
			},
			expectedError: &ErrSecretAlreadyExists{
				secretName: mockSecretName,
				parentErr:  mockAwsErr,
			},
		},

		"should return no error if successful": {
			inSecretName:   mockSecretName,
			inSecretString: mockSecretString,
			callMock: func(m *mocks.MockSecretsManagerAPI) {
				m.EXPECT().CreateSecret(&secretsmanager.CreateSecretInput{
					Name:         aws.String(mockSecretName),
					SecretString: aws.String(mockSecretString),
				}).Return(mockOutput, nil)
			},
			expectedError: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockSecretsManager := mocks.NewMockSecretsManagerAPI(ctrl)

			sm := SecretsManager{
				secretsManager: mockSecretsManager,
			}

			tc.callMock(mockSecretsManager)

			// WHEN
			_, err := sm.CreateSecret(tc.inSecretName, tc.inSecretString)

			// THEN
			require.Equal(t, tc.expectedError, err)
		})
	}
}
