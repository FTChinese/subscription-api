pipeline {
    agent any

    stages {
        stage('Build') {
            steps {
                echo 'Download config file'
                sh 'make config'
                echo 'Build superyard app'
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
                echo 'publish binary'
                sh 'make publish'
                echo 'restart app'
                sh 'make restart'
            }
        }
    }
}
