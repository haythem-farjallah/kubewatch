# kubewatch

`kubewatch` is a small Go CLI tool that watches Kubernetes pod changes in real time.

It connects to your Kubernetes cluster using your local kubeconfig and prints pod events in a clean table format.

## Features

- Watch pod changes in real time
- Filter by namespace
- Watch all namespaces
- Filter pods by label selector
- Colored event types
- Show useful pod details:
  - Ready containers
  - Pod status
  - Restart count
  - Node name
- Graceful shutdown with `Ctrl+C`

## Requirements

- Go
- kubectl
- Access to a Kubernetes cluster
- A valid kubeconfig file at `~/.kube/config`

## Installation

```bash
git clone https://github.com/haythem-farjallah/kubewatch.git
cd kubewatch
go mod tidy
```

## Usage

Watch pods in the default namespace:

```bash
go run .
```

Watch pods in a specific namespace:

```bash
go run . -n kube-system
```

```bash
go run . --namespace kube-system
```

Watch pods in all namespaces:

```bash
go run . -A
```

```bash
go run . --all-namespaces
```

Watch pods by label selector:

```bash
go run . -l app=nginx
```

```bash
go run . --selector app=nginx
```

Combine namespace and label filtering:

```bash
go run . -n dev -l app=api
```

Watch all namespaces with a label selector:

```bash
go run . -A -l app=nginx
```

## Example output

```text
TIME       EVENT        POD                  NAMESPACE    READY        STATUS       RESTARTS     NODE
18:42:10   ADDED        nginx-test           default      0/1          Pending      0            <none>
18:42:12   MODIFIED     nginx-test           default      1/1          Running      0            minikube
18:43:01   DELETED      nginx-test           default      1/1          Running      0            minikube
```

## Local testing

Create a test pod:

```bash
kubectl run nginx-test --image=nginx --labels app=nginx
```

Watch only pods with that label:

```bash
go run . -l app=nginx
```

Delete the test pod:

```bash
kubectl delete pod nginx-test
```

Test another namespace:

```bash
kubectl create namespace dev
kubectl run api-test --image=nginx -n dev --labels app=api
go run . -n dev -l app=api
kubectl delete pod api-test -n dev
```

## Roadmap

- Add Kubernetes Events watch mode
- Add warning-only event filtering
- Add object filtering with `--for pod/name`
- Add reconnect logic for long-running watches
- Add JSON output mode
- Refactor project structure into packages