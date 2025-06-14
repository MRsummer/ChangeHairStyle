#!/bin/bash

# 设置环境变量
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

# 清理旧的构建文件
rm -rf build
rm -f hair_style_service.zip

# 创建构建目录
mkdir -p build

# 编译
echo "编译中..."
go build -o build/hair_style_service api/handler.go

# 复制配置文件
cp -r config build/
cp template.yaml build/

# 创建启动脚本
cat > build/scf_bootstrap << 'EOF'
#!/bin/bash
./hair_style_service
EOF

# 设置执行权限
chmod +x build/scf_bootstrap
chmod +x build/hair_style_service

# 打包
echo "打包中..."
cd build
zip -r ../hair_style_service.zip ./*
cd ..

echo "打包完成：hair_style_service.zip" 