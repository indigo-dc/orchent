#!/usr/bin/groovy

@Library(['github.com/indigo-dc/jenkins-pipeline-library']) _

pipeline {
    agent {
        label 'go'
    }
    
    environment {
        dockerhub_repo = "indigodatacloud/orchent"
        dockerhub_image_id = ""
    }

    stages {
        stage('Code fetching') {
            steps {
                checkout scm
            }
        }

        stage('Style Analysis') {
            steps {
                sh '''
                GOFMT="/usr/local/go/bin/gofmt -s"
                bad_files=$(find . -name "*.go" | xargs $GOFMT -l)
                if [[ -n "${bad_files}" ]]; then
                    echo "!!! '$GOFMT' needs to be run on the following files: "
                    echo "${bad_files}"
                    exit 1
                fi
                '''            
            }
        }

        stage('Dependency check') {
            agent {
                label 'docker-build'
            }
            steps {
                    OWASPDependencyCheckRun("$WORKSPACE/orchent", project="orchent")
            }
            post {
                always {
                    OWASPDependencyCheckPublish()
                    HTMLReport('', 'dependency-check-report.html', 'OWASP Dependency Report')
                    deleteDir()
                }
            }
        }
        
        stage('Metrics gathering') {
            agent {
                label 'sloc'
            }
            steps {
                SLOCRun()
            }
            post {
                success {
                    SLOCPublish()
                }
            }
        }

        stage('DockerHub delivery') {
            when {
                anyOf {
                    branch 'master'
                    buildingTag()
                }
            }
            agent {
                label 'docker-build'
            }
            steps {
                checkout scm
                dir("$WORKSPACE/utils") {
                    sh 'echo "echo \\$ORCHENT_VERSION > /tmp/orchent_version" >> build_docker.sh'
                    sh "./build_docker.sh ${env.dockerhub_repo}"
                    script {
                        def orchent_version = readFile('/tmp/orchent_version').trim()
                        dockerhub_image_id = "${dockerhub_repo}:${orchent_version}"
                    }
                    sh 'rm -f /tmp/orchent_version'
                }
            }
            post {
                success {
                    echo "Pushing Docker image ${dockerhub_image_id}.."
                    DockerPush(dockerhub_image_id)
                }
                failure {
                    echo 'Docker image building failed, removing dangling images..'
                    DockerClean()
                }
                always {
                    cleanWs()
                }
            }
        }

        stage('Build RPM/DEB packages') {
            when {
                anyOf {
                    buildingTag()
                    branch 'master'
                }
            }
            parallel {
                stage('Build on Ubuntu16.04') {
                    agent {
                        label 'bubuntu16'
                    }
                    steps {
                        checkout scm
                        sh ''' 
                            echo 'Within build on Ubuntu16.04'
                            ./utils/build-pkg.sh
                            mkdir -p UBUNTU
                            BUILD_PATH="../orchent_build_env/"
                            echo $BUILD_PATH
                            find "$BUILD_PATH" -type f -name "orchent*.deb" -exec cp -v -t UBUNTU \'{}\' \';\' 
                        '''            
                    }
                    post {
                        success {
                            archiveArtifacts artifacts: '**/UBUNTU/*.deb'                        }
                    }
                }
                stage('Build on CentOS7') {
                    agent {
                        label 'bcentos7'
                    }
                    steps {
                        checkout scm
                        sh '''
                            echo 'Within build on CentOS7'
                            ./utils/build-pkg.sh
                            mkdir -p RPMS
                            BUILD_PATH="../orchent_build_env/"
                            find "$BUILD_PATH" -type f -name "orchent*.rpm" -exec cp -v -t RPMS \'{}\' \';\'
                        '''
                    }
                    post {
                        success {
                            archiveArtifacts artifacts: '**/RPMS/*.rpm'
                        }
                    }
                }
            }
        }
    }
}
