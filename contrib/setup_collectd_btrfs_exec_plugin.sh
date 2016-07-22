#! /usr/bin/sh -e

: "${COLLECTD_BTRFS_USERID:=nobody}"
: "${COLLECTD_BTRFS_INTERVAL:=15}"
: "${COLLECTD_BTRFS_SETUP:=true}"
: "${COLLECTD_BTRFS_SUDO:=false}"

if [ "$COLLECTD_BTRFS_SETUP" != 'true' ]; then
  echo '$COLLECTD_BTRFS_SETUP not set to true, skipping.'
  exit 0
fi

. "$RS_ATTACH_DIR/rs_distro.sh"

case "$RS_DISTRO" in
  debian|ubuntu)
    plugin_dir="/etc/collectd/conf"
  ;;
  centos|suse*|redhat*)
    plugin_dir="/etc/collectd.d"
  ;;
  *)
    echo 'Distro not supported, exiting.'
    exit 1
  ;;
esac

sudo mkdir -p "$plugin_dir"
exec_plugin_conf="$plugin_dir/exec-btrfs.conf"

sudo mkdir -p /usr/local/bin
sudo cp -f "$RS_ATTACH_DIR/exec-btrfs" /usr/local/bin/
sudo chmod +x /usr/local/bin/exec-btrfs

plugin_exec='"/usr/local/bin/exec-btrfs"'

if [ "$COLLECTD_BTRFS_SUDO" = 'true' ]; then
  plugin_exec='"sudo" "/usr/local/bin/exec-btrfs"'
fi

read -r -d '' conf <<EOF
LoadPlugin exec
<Plugin exec>
  #     userid                     plugin executable             plugin args
  Exec  "$COLLECTD_BTRFS_USERID"   $plugin_exec   "-H" "$SERVER_UUID" "-i" "$COLLECTD_BTRFS_INTERVAL" "$COLLECTD_BTRFS_EXEC_MOUNTPOINT"
</Plugin>
EOF

echo "$conf" | sudo tee $exec_plugin_conf

# so non-root user can read the mountpoint
if [ "$COLLECTD_BTRFS_EXEC_MOUNTPOINT" = '/var/lib/docker' ]; then
  echo 'Relaxing permissions on /var/lib/docker'
  sudo chmod 0755 /var/lib/docker
fi

echo 'Restarting collectd.'
if type service >/dev/null 2>&1; then
  sudo service collectd restart
elif type systemctl >/dev/null 2>&1; then
  sudo systemctl restart collectd
elif [ -e /etc/init.d/collectd ]; then
  sudo /etc/init.d/collectd restart;
fi

echo 'Done.'
