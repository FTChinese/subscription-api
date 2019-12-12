#!/usr/bing/env groovy

pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                sh 'make build'
                sh 'make downconfig'
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
                sh 'make upconfig'
                sh 'make publish'
                sh 'make restart'
            }
        }
    }
}