#!/usr/bing/env groovy

pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                echo 'Build subscription api production'
//                 sh 'make install-go'
                sh 'make build'
                archiveArtifacts artifacts: 'build/**/*', fingerprint: true
            }
        }
        stage('Deploy') {
            when {
                expression {
                    currentBuild.result == null || currentBuild.result == 'SUCCESS'
                }
            }
            steps {
                echo 'Sync config file'
                sh 'make config'
                echo 'Copy binary'
                sh 'make publish'
                sh 'make restart'
            }
        }
    }
}
