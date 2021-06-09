library 'ops-jenkins-libs@master'

pipeline {
  agent none
  stages {
    stage ('Prep') {
      agent {
        docker {
          args '--net=host'
          image "${env.WEAVE_CI}"
          label 'dind'
          registryCredentialsId "${env.WEAVEREGISTRYCREDS}"
          registryUrl "${env.WEAVEREGISTRY}"
        }
      }
      steps {
        script {
          weave = readYaml file: '.weave.yaml'
          sh '/buildprep'
        }
        stash name: 'buildenv', includes: "buildprep.env"
      }
    }
    stage('Build') {
      agent {
        docker {
          image "${env.WEAVEBUILDER}"
          registryUrl "${env.WEAVEREGISTRY}"
          registryCredentialsId "${env.WEAVEREGISTRYCREDS}"
          label 'dind'
        }
      }
      steps {
        withCredentials([file(credentialsId: 'weavelabbotkey', variable: 'SSH_KEY')]) {
          sh '''
          install -D -m 400 $SSH_KEY /root/.ssh/id_rsa
          /usr/local/bin/gobuilder
          '''
          stash name: 'bins', includes: "${weave.slug}"
        }
      }
    }
    stage('Package') {
      agent {
        docker {
          args "--privileged"
          label "dind"
          image "${env.WEAVE_CI}"
          registryCredentialsId "${env.WEAVEREGISTRYCREDS}"
          registryUrl "${env.WEAVEREGISTRY}"
        }
      }
      steps {
        unstash "bins"
        unstash "buildenv"
        withCredentials([file(credentialsId: 'build-secrets', variable: 'BUILDSECRETS'), file(credentialsId: 'weavelabbotkey', variable: 'FILE')]) {
          sh '''
              . ./buildprep.env
              /bin/cd-tools
          '''
        }
      }
    }
  }
  post {
    success { weaveCIResult "Success" }
    aborted { weaveCIResult "Canceled" }
    failure { weaveCIResult "Failure" }
  }
}
