def createPackage(String architecture) {
    script {
        sh """
        cd orchent
        mkdir -p build/usr/bin
        mv build/orchent-${architecture}-linux build/usr/bin/orchent
        fpm -s dir -t deb -n orchent -v 1.0.0 -C build/ \\
            -p orchent_1.0.0_${architecture}.deb .
        """
    }
}

pipeline {
  agent none
  stages {
        stage('Build') {
            agent {
                docker {
                  label 'jenkinsworker00'
                  image 'golang:1.16.15'
                  reuseNode true
                }
            }
            steps {
                script {
                  sh '''
                  ls -latr
                  #rm -rf $WORKSPACE/*
                  #git clone https://github.com/indigo-dc/orchent.git
                  
                  #cd orchent
                  mkdir build
                  export GOCACHE=$WORKSPACE/.cache/go-build
                  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags ' -w -extldflags "-static"' -ldflags "-B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \n')" -o build/orchent-amd64-linux orchent.go
                  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -a -ldflags ' -w -extldflags "-static"' -ldflags "-B 0x$(head -c20 /dev/urandom|od -An -tx1|tr -d ' \n')" -o build/orchent-arm64-linux orchent.go
                  CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -ldflags '-w -extldflags "-static"' -o build/orchent-amd64-darwin orchent.go
                  CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -a -ldflags '-w -extldflags "-static"' -o build/orchent-arm64-darwin orchent.go
                  '''
                }  
            }
        }
        
        stage('Package') {
            agent {
                docker {
                    label 'jenkinsworker00'
                    image 'marica/fpm:latest'
                    reuseNode true
                }
            }
            steps {
                
                    createPackage('amd64')
                    createPackage('arm64')
                
            }
        }
  }
}
