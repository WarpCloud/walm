@Library('jenkins-library') _

def label = UUID.randomUUID().toString()
def podTemplateYaml=kubernetesTemplate.kubernetesTemplate('172.16.1.99/transwarp/tdc-tos-jnlp-slave-k8s')

timestamps {
  properties([buildDiscarder(
          logRotator(artifactDaysToKeepStr: '', artifactNumToKeepStr: '', daysToKeepStr: '60', numToKeepStr: '100')),
              gitLabConnection('gitlab-172.16.1.41'),
              parameters([string(defaultValue: '', description: '', name: 'RELEASE_TAG')]),
              pipelineTriggers([])
  ])
  updateGitlabCommitStatus(name: 'ci-build', state: 'pending')
  podTemplate(label: label, yaml: podTemplateYaml) {
    node(label) { container('builder') {
      currentBuild.result = "SUCCESS"

      waitDocker {}

      stage('scm checkout') {
        checkout(scm)
      }
      updateGitlabCommitStatus(name: 'ci-build', state: 'running')

      withEnv([
              'DOCKER_HOST=unix:///var/run/docker.sock',
              'DOCKER_REPO=172.16.1.99',
              'COMPONENT_NAME=walm',
              'DOCKER_PROD_NS=gold',
      ]) {

        try {
          withCredentials([
             usernamePassword(
                     credentialsId: 'harbor',
                     passwordVariable: 'DOCKER_PASSWD',
                     usernameVariable: 'DOCKER_USER')
          ]) {
              stage('release build') {
                  sh """#!/bin/bash -ex
                    docker login -u \$DOCKER_USER -p \$DOCKER_PASSWD \$DOCKER_REPO
                    REV=\$(git rev-parse HEAD)
                    export DOCKER_IMG_NAME=\$DOCKER_REPO/\$DOCKER_PROD_NS/\$COMPONENT_NAME:${env.BRANCH_NAME}
                    docker build --label CODE_REVISION=\${REV} \
                      --label BRANCH=$env.BRANCH_NAME \
                      --label COMPILE_DATE=\$(date +%Y%m%d-%H%M%S) \
                      -t \$DOCKER_IMG_NAME -f Dockerfile .
                    docker push \$DOCKER_IMG_NAME
                  """
              }
          }
          updateGitlabCommitStatus(name: 'ci-build', state: 'success')
        } catch (e) {
          currentBuild.result = "FAILED"
          updateGitlabCommitStatus(name: 'ci-build', state: 'failed')
          echo 'Err: Incremental Build failed with Error: ' + e.toString()
          throw e
        } finally {
          sendMail2 {
              emailRecipients = "tosdev@transwarp.io"
              attachLog = false
          }
        }
      }
    }}
  }
}
