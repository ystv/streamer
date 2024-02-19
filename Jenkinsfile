String registryEndpoint = 'registry.comp.ystv.co.uk'

def branch = env.BRANCH_NAME.replaceAll("/", "_")
def image
String proceed = "yes"
String serverImageName = "ystv/streamer/server:${branch}-${env.BUILD_ID}"
String forwarderImageName = "ystv/streamer/forwarder:${branch}-${env.BUILD_ID}"
String recorderImageName = "ystv/streamer/recorder:${branch}-${env.BUILD_ID}"
def productionBuild = env.TAG_NAME ==~ /v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)/

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
                  serverImage = docker.build(serverImageName, "--build-arg STREAMER_VERSION_ARG=${env.BRANCH_NAME}-${env.BUILD_ID} --no-cache .")
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
                  forwarderImage = docker.build(forwarderImageName, "--build-arg STREAMER_VERSION_ARG=${env.BRANCH_NAME}-${env.BUILD_ID} --no-cache .")
                }
              }
            }
          }
        }
        stage('Build Recorder') {
          environment {
            STREAMER_RECORDER_USER_UID = credentials('streamer-recorder-user-uid')
            STREAMER_RECORDER_USER_GID = credentials('streamer-recorder-user-gid')
            STREAMER_RECORDER_USER_GROUP_NAME = credentials('streamer-recorder-user-group-name')
          }
          steps {
            script {
              dir("recorder") {
                docker.withRegistry('https://' + registryEndpoint, 'docker-registry') {
                  recorderImage = docker.build(recorderImageName, "--build-arg STREAMER_VERSION_ARG=${env.BRANCH_NAME}-${env.BUILD_ID} --build-arg STREAMER_RECORDER_USER_UID=${STREAMER_RECORDER_USER_UID} --build-arg STREAMER_RECORDER_USER_GID=${STREAMER_RECORDER_USER_GID} --build-arg STREAMER_RECORDER_USER_GROUP_NAME=${STREAMER_RECORDER_USER_GROUP_NAME} --no-cache .")
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

    stage('Checking existing streams') {
      steps {
        script {
          String url = "https://streamer."
          if (!productionBuild) {
            url += "dev."
          }
          url += "ystv.co.uk/activeStreams"
          final def (String response, String tempCode) =
              sh(script: "curl -s -w '~~~%{response_code}' $url", returnStdout: true)
                  .trim()
                  .tokenize("~~~")
          int code = Integer.parseInt(tempCode)

          echo "HTTP response status code: $code"
          echo "HTTP response: $response"

          if (code == 200) {
            if (response.contains("streams")) {
              tempStreams = sh(script: "echo '$response' | jq -M '.streams'", returnStdout: true).trim()
              int streams = Integer.parseInt(tempStreams)
              if (streams > 0) {
                echo "Pre-existing active streams: $streams, not deploying"
                proceed = "no"
              } else {
                echo "No pre-existing active streams, deploying"
              }
            } else {
              echo "Streamer not currently running, deploying"
            }
          } else {
            echo "Invalid HTTP response code: $code, deploying"
          }
        }
      }
    }

    stage('Deploy') {
      when {
        expression { proceed == "yes" }
      }
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
            expression { return productionBuild }
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