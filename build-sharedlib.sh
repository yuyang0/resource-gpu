CURDIR=`pwd`

cd /tmp
git clone https://github.com/projecteru2/core.git
cd core
go mod tidy
go mod vendor

tee -a go.mod <<EOF
require github.com/yuyang0/resource-gpu v0.0.0-00010101000000-000000000000
replace github.com/yuyang0/resource-gpu => $CURDIR
EOF

go build -buildmode=plugin -mod=readonly github.com/yuyang0/resource-gpu
cp resource-gpu.so $CURDIR
rm -rf /tmp/core