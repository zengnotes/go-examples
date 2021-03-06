
## Setup

You need an oci8.pc file and a PKG_CONFIG_PATH that points to a directory that
includes it before
you can install the package. This also means that the `github.com/rjeczalik/pkgconfig/cmd/pkg-config`
package must also be installed in your go path.

PKG_CONFIG_PATH should be an absolute path, and should not end with a trailing slash.

Once the above is done, the rana/ora package can be installed:

<pre>
go get gopkg.in/rana/ora.v3
</pre>

If behind a proxy, you will probably need to set your http_proxy and
https_proxy environment variabled before you go get the package. I also
noticed an absolute path is needed for PKG_CONFIG_PATH.

For the blob sample first create a table like this:

<pre>
create table blob_sample (
    id integer,
    payload blob
);
</pre>

