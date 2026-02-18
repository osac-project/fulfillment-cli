# Fulfillment CLI

The Fulfillment CLI is a command-line tool for managing infrastructure resources through the
fulfillment API. It is designed to work with the
[fulfillment-service](https://github.com/osac-project/fulfillment-service) using the API defined in
[fulfillment-api](https://github.com/osac-project/fulfillment-api). The CLI provides a simple,
intuitive interface for working with clusters, virtual machines, hosts, and other infrastructure
components in your environment.

With this tool, you can create and manage _OpenShift_ clusters, provision virtual machines, manage
host pools, and perform other infrastructure operations directly from your terminal. The CLI
communicates with the fulfillment service using _gRPC_ and supports secure authentication through
_OAuth_.

## Installation

The easiest way to install this CLI is to download a pre-built binary from the
[GitHub releases page](https://github.com/osac-project/fulfillment-cli/releases). Choose the release
that matches your operating system and architecture.

Once you've downloaded the binary, make it executable and move it to a directory in your system's
PATH. For example, on Linux or macOS:

```bash
$ chmod +x fulfillment-cli
$ sudo mv fulfillment-cli /usr/local/bin/
```

You can verify the installation by running `fulfillment-cli version` to display the version
information.

## Getting started

Before you can use the CLI to manage resources, you need to authenticate with the Fulfillment
service. The `login` command handles this process and saves your credentials for future use.

To log in, provide the address of the fulfillment service:

```bash
$ fulfillment-cli login api.example.com:443
```

The CLI will guide you through the authentication process, which typically involves opening a
browser window to complete the _OAuth_ flow. Once authenticated, your credentials are stored
locally and automatically used for subsequent commands.

## Working with templates

Templates define the blueprint for creating infrastructure objects such as _OpenShift_ clusters
and virtual machines. The principles described in this section apply to all types of templates,
but we'll use cluster templates as our example. Before creating any object, you'll want to see
what templates are available in your environment.

To list all available templates of a particular type, use the `get` command. For example, to list
cluster templates:

```bash
$ fulfillment-cli get clustertemplates
```

This displays a table showing the available templates along with their key properties. For cluster
templates, the output might look like this:

```
ID              NAME  TITLE                 DESCRIPTION
ocp_4_17_small  -     OpenShift 4.17 small  OpenShift 4.17 with `small` instances as worker nodes.
```

If you want more detailed information about a specific template, you can get the YAML or JSON
representation. This works for any template type:

```bash
$ fulfillment-cli get clustertemplate -o yaml
```

That will print this:

```yaml
'@type': type.googleapis.com/fulfillment.v1.ClusterTemplate
description: OpenShift 4.17 with `small` instances as worker nodes.
id: ocp_4_17_small
metadata:
  creation_timestamp: "2025-11-04T10:54:41.666545Z"
title: OpenShift 4.17 small
parameters:
  - default:
      '@type': type.googleapis.com/google.protobuf.BoolValue
      value: true
    name: my_bool
    type: type.googleapis.com/google.protobuf.BoolValue
  - default:
      '@type': type.googleapis.com/google.protobuf.Int32Value
      value: 42
    name: my_int
    type: type.googleapis.com/google.protobuf.Int32Value
  - default:
      '@type': type.googleapis.com/google.protobuf.StringValue
      value: my_value
    name: my_string
    type: type.googleapis.com/google.protobuf.StringValue
```

The template details show all the configuration parameters, including default values and types.
These parameters can be customized when creating objects from the template.

## Creating objects

The CLI supports creating various types of infrastructure objects including clusters, virtual
machines, and other resources. The principles described in this section apply to all object types,
but we'll use clusters as our example.

Once you've identified the template you want to use, creating an object is straightforward. The
`create` command accepts the object type, template identifier, and any required parameters. For
example, to create a cluster:

```bash
$ fulfillment-cli create cluster --template
Created cluster '0ad55e76-fefb-451d-a812-21ce39c3ed06'.
```

Templates accept additional parameters to customize the object configuration. You can pass these
parameters using the `--parameter` or `-p` flag:

```bash
$ fulfillment-cli create cluster --template  ocp_4_17_small --name my-cluster \
-p my_bool=true \
-p my_int=43 \
-p my_value=whatever
```

After creating an object, you can monitor its status with the `get` command. The same pattern
works for any object type:

```bash
$ fulfillment-cli get cluster Created cluster '019a4f3c-77fe-77db-9ef4-d4b7d141499e'.
```

To see detailed information about a specific object, use the describe command:

```bash
$ fulfillment-cli describe cluster 0ad55e76-fefb-451d-a812-21ce39c3ed06
```

Some object types have additional operations specific to them. For example, once a cluster is
ready, you can retrieve its kubeconfig file to start using it with kubectl:

```bash
$ fulfillment-cli get kubeconfig 0ad55e76-fefb-451d-a812-21ce39c3ed06 > kubeconfig.yaml
export KUBECONFIG=kubeconfig.yaml
kubectl get nodes
```

## Additional commands

Beyond creating and viewing objects, the CLI provides several other useful commands for managing
your infrastructure. The `edit` command allows you to modify existing objects interactively, and
the `delete` command removes objects you no longer need. These commands work with all object types
using the same consistent interface.

For a complete list of available commands, object types, and their options, run
`fulfillment-cli --help`. Each command also has its own help text available with
`fulfillment-cli <command> --help`.

## Configuration

The CLI stores its configuration in your home directory under `.config/fulfillment-cli/config`.
This includes your authentication credentials and connection details. You can log out and remove
these credentials at any time with the `logout` command:

```bash
$ fulfillment-cli logout
```

## Logging

By default, the CLI writes log files to your system's cache directory (typically
`~/.cache/fulfillment-cli/fulfillment-cli.log`) to keep the output clean. If you need to see
detailed logs for troubleshooting, you can increase the logging level with the
`--log-level=debug` flag and write the log to the console with `--log-file=stdout`,
for example:

```bash
$ fulfillment-cli --log-level debug get clusters
```
