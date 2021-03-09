#!/usr/bing/env groovy

pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                sh 'make install-go'
                sh 'make amd64'
                archiveArtifacts artifacts: 'build/*', fingerprint: true
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
                sh 'make publish'
                sh 'make restart'
            }
        }
    }
}
