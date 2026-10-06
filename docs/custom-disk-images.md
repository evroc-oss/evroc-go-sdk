# Custom disk images

Use custom images to ship preinstalled applications, security baselines,
monitoring agents, or cached dependencies. Moving setup into the image build
can make new VMs start faster and more consistently. Kubernetes node images
are one use case; the same API works for general-purpose customer workloads.

Custom disk images register an image already uploaded to evroc object storage.
The SDK exposes the Compute v1beta2 `customDiskImages` API and uses the client's
configured project and region. It does not build, upload, scan or modify image
contents. The platform custom disk image API must be available before these
calls can succeed.

## Build an image (Ubuntu example)

On a Linux build host, install `qemu-utils` and `libguestfs-tools`. Download an
amd64 Ubuntu 24.04 cloud image from [Canonical](https://cloud-images.ubuntu.com/noble/),
verify its signed checksum, and save it as `ubuntu-base.img`. Use a fresh cloud
image rather than a disk from a configured VM. The example assumes qcow2;
confirm with `qemu-img info ubuntu-base.img` before proceeding.

```bash
set -e
image=my-image-v1.qcow2
test ! -e "$image"
qemu-img convert -f qcow2 -O qcow2 ubuntu-base.img "$image"

# Install packages now instead of downloading them during first boot.
virt-customize --format qcow2 -a "$image" \
  --install curl,jq \
  --run-command 'apt-get clean'

# Remove identity and build state so each VM initializes independently.
virt-sysprep --format qcow2 -a "$image" \
  --operations ssh-hostkeys,ssh-userdir,bash-history,logfiles,tmp-files
virt-customize --format qcow2 -a "$image" \
  --run-command 'cloud-init clean --logs --machine-id --seed'
qemu-img check -f qcow2 "$image"
sha256sum "$image"
```

Replace `curl,jq` with your application's packages; use `virt-customize --run
provision.sh` for more setup. Keep cloud-init enabled for SSH keys and user-data.
Do not bake credentials or cluster identity into the image. For Kubernetes,
preinstall binaries or cache container images, but leave cluster joining and
CSI deployment/configuration to bootstrap. Pin the base image and package
versions in a repeatable build pipeline.

Upload `my-image-v1.qcow2` to your evroc bucket as `images/my-image.qcow2`, then
register it below using the uploaded object's version. The default disk size
must cover the image's **virtual size**, not its compressed file size. Boot-test
a new VM for cloud-init, networking, SSH and filesystem growth before publishing
the image. Package preinstallation can shorten first boot, but larger images
also take longer to import.

Tool references: [virt-customize](https://libguestfs.org/virt-customize.1.html),
[virt-sysprep](https://libguestfs.org/virt-sysprep.1.html),
[cloud-init clean](https://docs.cloud-init.io/en/latest/reference/cli.html#clean).

## Register an image

Upload a supported image to a bucket in the target project and create a bucket
service account authorized to read it. Supply fully qualified bucket and account
references; the SDK does not infer those from names or carry S3 credentials in
the registration request.

```go
image, err := compute.NewCustomDiskImageBuilder(
    "my-image-v1",
    "/storage/projects/my-project/regions/se-sto/buckets/images",
    "/storage/projects/my-project/regions/se-sto/bucketServiceAccounts/image-reader",
    "images/my-image.qcow2",
).
    WithDefaultDiskSizeGB(50).
    WithObjectVersion("s3-object-version-id").
    WithDescription("Preconfigured OS image").
    WithOSName("ubuntu").
    WithOSVersion("24.04").
    WithImageVersion("release-1").
    Create(ctx, client.Compute().CustomDiskImages())
if err != nil {
    return err
}
image, err = client.Compute().CustomDiskImages().WaitForReady(
    ctx, image.Metadata.Id, 2*time.Minute,
)
if err != nil {
    return err
}
```

The builder defaults to `amd64`, the architecture currently advertised by this
API. A default disk size is required; choose at least the image's virtual disk
size. Other constraints are validated by the platform.

`WithObjectVersion` pins the stored object's version. `WithImageVersion` is only
descriptive metadata. Without an object version, new disks use the latest object
at that path, even though the registration's source fields are immutable.

Registration readiness confirms the platform is ready to use the registration;
it does not establish that the object is a valid or bootable image. Download,
conversion and disk readiness are handled when creating the disk.

## Create a boot disk

```go
req := compute.NewDiskBuilder("boot-disk").
    WithCustomImage(client.Compute().CustomDiskImageRef("my-image-v1")).
    WithZone("a").
    WithSizeGB(50).
    Build()
disk, err := client.Compute().Disks().Create(ctx, req)
if err != nil {
    return err
}
disk, err = client.Compute().Disks().WaitForReady(ctx, disk.Metadata.Id, 5*time.Minute)
if err != nil {
    return err
}
```

You can also pass `image.Ref()` when a fetched image has project/region metadata.
`CustomDiskImageRef` is a typed fully qualified reference, not a bare image name.
Reference helpers do not grant access: project scope and permissions are enforced
by the API. Consumers such as Cluster API can use this builder without assembling
resource paths or editing the raw disk request themselves.

Omit `WithSizeGB` to let the platform use the image's default disk size. Standard
images still use `WithImage("ubuntu.24-04.1")`; snapshots use `WithSnapshot(ref)`.
`WithCustomImage` replaces the previous source; passing an empty custom reference
produces a blank disk. Existing standard-image and snapshot behavior is preserved.

## Manage registrations

Use `CustomDiskImages().List(ctx, filter.WithLabelSelector("team=infra"))` or
`Get(ctx, name)` to find registrations. The usual `Labels()` helper is supported.
Patch descriptive metadata without resending immutable source fields:

```go
_, err := client.Compute().CustomDiskImages().Patch(ctx, "my-image-v1",
    map[string]interface{}{
        "spec": map[string]interface{}{"description": "Updated description"},
    },
)
```

Use a new registration and object version for a new source image. Architecture,
default disk size and source are immutable under the platform API. `Delete` and
`WaitForDeleted` manage the registration only; they do not delete bucket objects,
bucket accounts, existing disks or VMs. Stop creating disks from an image and
follow platform reference/deletion rules before retiring it.

See the [runnable example](../examples/custom-images/main.go). Image contents,
bootstrap compatibility and cluster add-ons are choices for the image publisher
and consuming application, not policies of the SDK.
