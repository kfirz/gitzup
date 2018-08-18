# Manifests

Gitzup manifests are essentially a list of resources, where for each resource a declaration of desired state is provided. By *desired state* we mean (in broad strokes) **_how the resource should be once the manifest is fully applied_**.

By allowing you - the manifest author - to concentrate on how the resources should _be_ (eg. _my cluster should have 4 nodes_ or _my VM should be assigned the "us-east1-static-ip" address_) rather on commands and instructions, Gitzup frees you from the mundane DevOps tasks & know-how required to reach such state. You don't have to know **_how_** to attach static IP addresses to VMs; you just need to know you want to do it!

In other words, Gitzup provides a way to transition your deployment topology from a **_current state_** to a **_desired state_** in a declarative way.

## Resources

Since there are endless types of resources (Gitzup tends to see everything as a resource: static IPs, VMs, clusters, nodes, Kubernetes objects, disks, GitHub Releases, DNS records, etc), Gitzup needs an easy & extensible way to support custom resources that Gitzup cannot and will-not be able to support upfront. Enter (drum roll please!) **Docker**!

Each resource in Gitzup is essentially a Docker image written explicitly to provide support for a resource or a set of resource types to Gitzup. You can think of it as an adapter between Gitzup's generic resource model and the resource-specific domain. Therefor, each resource declaration in a manifest will always refer to a Docker image in its `type` property.

For example:
```yaml
version: 1
resources:
  us-east1-ip:
    type: gitzup/gcp-ip-address
```

Most resources will also require some configuration before they can be applied by Gitzup. You provide such configuration through the `config` property, like this:

```yaml
version: 1
resources:
  us-east1-ip:
    type: gitzup/gcp-ip-address
    config:
      project_id: western-lights
      region: us-east1
```

When such a manifest will be applied, Gitzup will check if a reserved static IP address in project `western-lights` exists in region `us-east1`; if not, one will be created.
