# Maintainer: Paulo Ruan PauloRuan_30@outlook.com
pkgname=tome
pkgver=0.1.0
pkgrel=1
pkgdesc="Interactive TUI PDF bookshelf manager and reader"
arch=('x86_64')
url="https://github.com/PauloRuan30/tome"
license=('GPL-3.0-only')
depends=('glibc')
makedepends=('go' 'git')
source=("git+https://github.com/PauloRuan30/tome.git")
sha256sums=('SKIP')

build() {
  cd "$pkgname"
  make build
}

package() {
  cd "$pkgname"
  make install INSTALL_DIR="$pkgdir/usr/bin"
}