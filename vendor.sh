#!/usr/bin/env bash
set -e

# Downloads dependencies into vendor/ directory
mkdir -p vendor
cd vendor

clone() {
	vcs=$1
	pkg=$2
	rev=$3

	pkg_url=https://$pkg
	target_dir=src/$pkg

	echo -n "$pkg @ $rev: "

	if [ -d $target_dir ]; then
		echo -n 'rm old, '
		rm -fr $target_dir
	fi

	echo -n 'clone, '
	case $vcs in
		git)
			git clone --quiet --no-checkout $pkg_url $target_dir
			( cd $target_dir && git reset --quiet --hard $rev )
			;;
		hg)
			hg clone --quiet --updaterev $rev $pkg_url $target_dir
			;;
	esac

	echo -n 'rm VCS, '
	( cd $target_dir && rm -rf .{git,hg} )

	echo done
}

clone git github.com/opencontainers/runc v0.0.2
clone git github.com/codegangsta/cli v1.2.0
clone git github.com/Sirupsen/logrus v0.8.2
