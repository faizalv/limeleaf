#!/usr/bin/env bash
set -euo pipefail

PG_VERSION="${PG_VERSION:-16.4}"
PGVECTOR_VERSION="${PGVECTOR_VERSION:-0.7.0}"
OUTPUT_DIR="${OUTPUT_DIR:-$(pwd)/dist}"

WORKDIR="$(mktemp -d)"
PREFIX="${WORKDIR}/install"
trap 'rm -rf "$WORKDIR"' EXIT

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)        ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac
PLATFORM="${OS}-${ARCH}"

if command -v nproc &>/dev/null; then
    JOBS="$(nproc)"
else
    JOBS="$(sysctl -n hw.ncpu)"
fi

echo "==> Building Postgres ${PG_VERSION} + pgvector ${PGVECTOR_VERSION} for ${PLATFORM}"

# --- Postgres ---

PG_TARBALL="postgresql-${PG_VERSION}.tar.bz2"
PG_URL="https://ftp.postgresql.org/pub/source/v${PG_VERSION}/${PG_TARBALL}"

echo "==> Downloading PostgreSQL ${PG_VERSION}"
curl -fSL -o "${WORKDIR}/${PG_TARBALL}" "${PG_URL}"
tar xf "${WORKDIR}/${PG_TARBALL}" -C "${WORKDIR}"

echo "==> Configuring PostgreSQL"
cd "${WORKDIR}/postgresql-${PG_VERSION}"

CONFIGURE_FLAGS=(
    --prefix="${PREFIX}"
    --without-readline
    --without-icu
    --disable-nls
)

if [[ "$OS" == "linux" ]]; then
    CONFIGURE_FLAGS+=(--without-zlib)
fi

./configure "${CONFIGURE_FLAGS[@]}"

echo "==> Building PostgreSQL"
make -j"${JOBS}"
make install

# --- pgvector ---

PGVECTOR_URL="https://github.com/pgvector/pgvector/archive/refs/tags/v${PGVECTOR_VERSION}.tar.gz"

echo "==> Downloading pgvector ${PGVECTOR_VERSION}"
curl -fSL -o "${WORKDIR}/pgvector.tar.gz" "${PGVECTOR_URL}"
tar xf "${WORKDIR}/pgvector.tar.gz" -C "${WORKDIR}"

echo "==> Building pgvector"
cd "${WORKDIR}/pgvector-${PGVECTOR_VERSION}"
make PG_CONFIG="${PREFIX}/bin/pg_config" -j"${JOBS}"
make PG_CONFIG="${PREFIX}/bin/pg_config" install

# --- Strip ---

echo "==> Stripping binaries"
if [[ "$OS" == "darwin" ]]; then
    STRIP_CMD="strip -x"
else
    STRIP_CMD="strip --strip-unneeded"
fi

find "${PREFIX}/bin" -type f -perm -u+x -exec $STRIP_CMD {} \; 2>/dev/null || true
find "${PREFIX}/lib" -type f \( -name "*.so" -o -name "*.dylib" \) -exec $STRIP_CMD {} \; 2>/dev/null || true

# --- Prune ---

echo "==> Removing unnecessary files"
rm -rf "${PREFIX}/include"
rm -rf "${PREFIX}/share/doc"
rm -rf "${PREFIX}/share/man"
rm -rf "${PREFIX}/lib/pgxs"
rm -rf "${PREFIX}/lib/pkgconfig"

KEEP_BINS="postgres initdb pg_ctl pg_isready psql pg_config"
for bin in "${PREFIX}/bin/"*; do
    name="$(basename "$bin")"
    keep=false
    for k in $KEEP_BINS; do
        if [[ "$name" == "$k" ]]; then
            keep=true
            break
        fi
    done
    if ! $keep; then
        rm -f "$bin"
    fi
done

# --- Package ---

echo "==> Creating tarball"
mkdir -p "${OUTPUT_DIR}"
TARBALL="${OUTPUT_DIR}/${PLATFORM}.tar.gz"

cd "${PREFIX}"
tar czf "${TARBALL}" .

if command -v sha256sum &>/dev/null; then
    sha256sum "${TARBALL}" | awk '{print $1}' > "${TARBALL}.sha256"
else
    shasum -a 256 "${TARBALL}" | awk '{print $1}' > "${TARBALL}.sha256"
fi

echo "==> Done: ${TARBALL}"
echo "    SHA256: $(cat "${TARBALL}.sha256")"
