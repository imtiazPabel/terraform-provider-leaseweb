#!groovy

lswci([node: 'docker', mattermost: 'bare-metal-cicd']) {
    name = env.CHANGE_BRANCH ? env.CHANGE_BRANCH.toLowerCase().replace("/", "-") : env.BRANCH_NAME.toLowerCase().replace("/", "-")

    image = docker.build("${name}-dev", "--target godev .")
    // TODO make unpivileged in future
    image.inside("--privileged -u root") {
        stage("Lint") {
            sh "make ci"
        }

        stage("Test build") {
            sh "make build"
        }

        if (env.BRANCH_NAME == 'master') {
            stage("Build release") {
                sh "make release"
            }

            stage("Publish artifacts") {
                step([$class: 'ArtifactArchiver', artifacts: 'dist/*', fingerprint: true])
            }
        }
    }
}
