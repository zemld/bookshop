#!/usr/bin/env bash
set -euo pipefail

if [[ $# != 2 || ! $2 =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
    echo "Usage: $0 {build|migrate|deploy} vMAJOR.MINOR.PATCH" >&2
    exit 2
fi

action=$1
version=$2
profile=${MINIKUBE_PROFILE:-minikube}
cd "$(git rev-parse --show-toplevel)"
revision=$(git rev-parse --verify "refs/tags/${version}^{commit}")

case "$action" in
    build)
        source_dir=$(mktemp -d)
        trap 'rm -rf "$source_dir"' EXIT
        git archive "$revision" | tar -x -C "$source_dir"
        for service in backend frontend; do
            docker build --build-arg "VERSION=$version" --build-arg "REVISION=$revision" \
                -t "bookshop-$service:$version" -f "$source_dir/$service/deploy/Dockerfile" "$source_dir/$service"
        done
        docker build -t "bookshop-migrate:$version" \
            -f "$source_dir/backend/deploy/Dockerfile.migrate" "$source_dir/backend"
        docker save "bookshop-backend:$version" "bookshop-frontend:$version" "bookshop-migrate:$version" |
            minikube -p "$profile" image load -
        ;;
    migrate)
        job="migrate-${version//./-}"
        git show "$revision:infra/minikube/migrate.yaml" |
            kubectl --context "$profile" set image --local -f - "migrate=bookshop-migrate:$version" -o json |
            jq --arg job "$job" '.metadata.name = $job' |
            kubectl --context "$profile" -n bookshop apply -f -
        kubectl --context "$profile" -n bookshop wait --for=condition=complete "job/$job" --timeout=180s
        ;;
    deploy)
        for service in backend frontend; do
            git show "$revision:infra/minikube/application/$service-deployment.yaml" |
                kubectl --context "$profile" set image --local -f - "$service=bookshop-$service:$version" -o yaml |
                kubectl --context "$profile" -n bookshop apply -f -
            kubectl --context "$profile" -n bookshop rollout status "deployment/$service" --timeout=180s
        done
        ;;
    *)
        echo "Unknown release action: $action" >&2
        exit 2
        ;;
esac
