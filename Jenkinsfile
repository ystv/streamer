String registryEndpoint = 'registry.comp.ystv.co.uk'

def branch = env.BRANCH_NAME.replaceAll("/", "_")
def image
String serverImageName = "ystv/streamer/server:${branch}-${env.BUILD_ID}"
String forwarderImageName = "ystv/streamer/forwarder:${branch}-${env.BUILD_ID}"
String recorderImageName = "ystv/streamer/recorder:${branch}-${env.BUILD_ID}"

pipeline {
  agent {
    label 'docker'
  }

  environment {
    DOCKER_BUILDKIT = '1'
  }

  stages {
    stage('Build images') {
      steps {
        script {
          dir("server") {
            docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
              serverImage = docker.build(serverImageName, ".")
            }
          }
        }
        script {
          dir("forwarder") {
            docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
              forwarderImage = docker.build(forwarderImageName, ".")
            }
          }
        }
        script {
          dir("recorder") {
            docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
              recorderImage = docker.build(recorderImageName, ".")
            }
          }
        }
      }
    }

    stage('Push images to registry') {
      steps {
        script {
          docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
            serverImage.push()
            if (env.BRANCH_IS_PRIMARY) {
              serverImage.push('latest')
            }
          }
        }
        script {
          docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
            forwarderImage.push()
            if (env.BRANCH_IS_PRIMARY) {
              forwarderImage.push('latest')
            }
          }
        }
        script {
          docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
            recorderImage.push()
            if (env.BRANCH_IS_PRIMARY) {
              recorderImage.push('latest')
            }
          }
        }
      }
    }

    stage('Deploy') {
      stages {
        stage('Development') {
          when {
            expression { env.BRANCH_IS_PRIMARY }
          }
          steps {
            build(job: 'Deploy Nomad Job', parameters: [
              string(name: 'JOB_FILE', value: 'streamer-dev.nomad'),
              text(name: 'TAG_REPLACEMENTS', value: "${registryEndpoint}/${serverImageName} ${registryEndpoint}/${forwarderImageName} ${registryEndpoint}/${recorderImageName}")
            ])
          }
        }

        stage('Production') {
          when {
            // Checking if it is semantic version release.
            expression { return env.TAG_NAME ==~ /v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)/ }
          }
          steps {
            build(job: 'Deploy Nomad Job', parameters: [
              string(name: 'JOB_FILE', value: 'streamer-prod.nomad'),
              text(name: 'TAG_REPLACEMENTS', value: "${registryEndpoint}/${serverImageName} ${registryEndpoint}/${forwarderImageName} ${registryEndpoint}/${recorderImageName}")
            ])
          }
        }
      }
    }
  }
}