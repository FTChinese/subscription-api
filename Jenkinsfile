#!/usr/bing/env groovy

pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                sh 'make build-api'
                archiveArtifacts artifacts: 'build/linux/*', fingerprint: true
            }
        }
        stage('Deploy') {
            when {
                expression {
                    currentBuild.result == null || currentBuild.result == 'SUCCESS'
                }
            }
            steps {
                sh 'make config'
                sh 'make publish-api'
                sh 'make restart-api'
            }
        }
    }
}
