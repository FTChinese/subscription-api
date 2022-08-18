pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                sh 'make config'
                sh 'make version'
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
                sh 'make publish'
                sh 'make restart'
            }
        }
    }
}
