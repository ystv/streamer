String registryEndpoint = 'registry.comp.ystv.co.uk'

def branch = env.BRANCH_NAME.replaceAll("/", "_")
def image
String proceed = "yes"
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
      parallel {
        stage('Build Server') {
          steps {
            script {
              dir("server") {
                docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
                  serverImage = docker.build(serverImageName, "--no-cache .")
                }
              }
            }
          }
        }
        stage('Build Forwarder') {
          steps {
            script {
              dir("forwarder") {
                docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
                  forwarderImage = docker.build(forwarderImageName, "--no-cache .")
                }
              }
            }
          }
        }
        stage('Build Recorder') {
          steps {
            script {
              dir("recorder") {
                docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
                  recorderImage = docker.build(recorderImageName, "--no-cache .")
                }
              }
            }
          }
        }
      }
    }

    stage('Push images to registry') {
      parallel {
        stage('Push Server image to registry') {
          steps {
            script {
              docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
                serverImage.push()
                if (env.BRANCH_IS_PRIMARY) {
                  serverImage.push('latest')
                }
              }
            }
          }
        }
        stage('Push Forwarder image to registry') {
          steps {
            script {
              docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
                forwarderImage.push()
                if (env.BRANCH_IS_PRIMARY) {
                  forwarderImage.push('latest')
                }
              }
            }
          }
        }
        stage('Push Recorder image to registry') {
          steps {
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
      }
    }

    stage('Deploy') {
      stages {
        stage('Checking existing') {
          steps {
            script {
              final String url = "https://streamer.dev.ystv.co.uk/activeStreams"
              final def (String response, int code) =
                  sh(script: "curl -s $url", returnStdout: true)
                      .trim()
                      .tokenize("\n")

              echo "HTTP response status code: $code"
              echo "HTTP response: $response"

              if (code == 200) {
                  def streams = sh(script: "echo '$response' | jq -M '.streams'", returnStdout: true)
                  if (streams > 0) {
                    proceed = "no"
                  }
              }
            }
          }
        }
        stage('Development') {
          when {
            expression { env.BRANCH_IS_PRIMARY && proceed == "yes" }
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
            expression { return env.TAG_NAME ==~ /v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)/ && proceed == "yes" }
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