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
                configFileProvider([configFile('8dab81ba-ecfe-4716-9201-33121b18c470', variable: 'API_CONFIG')]) {
                    sh 'scp ${API_CONFIG} node11:/home/node/go/api.toml'
                }
                sh 'make publish'
                sh 'make restart'
            }
        }
    }
}