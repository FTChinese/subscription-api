#!/usr/bing/env groovy

pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                sh 'make linux'
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
                sh 'make publish'
                sh 'make restart'
            }
        }
    }
}