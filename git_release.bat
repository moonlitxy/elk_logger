@echo off
chcp 65001
setlocal enabledelayedexpansion

echo 获取版本号参数
set VERSION=%1

echo 检查版本号是否为空
if "%VERSION%"=="" (
    echo 请指定版本号，例如: .\git_release.bat v1.0.0
    exit /b
)


echo 检查 Git 状态...
git status

echo 添加所有修改
echo 添加所有修改的文件到暂存区...
git add .

echo 提交修改
echo 提交修改: 更新到 %VERSION%
git commit -m "更新到 %VERSION%"

echo 推送 main 分支
echo 推送代码到远程主分支...
git push origin main

echo 创建 tag
echo 创建新的 tag: %VERSION%
git tag %VERSION%

echo 推送 tag 到远程
echo 推送 tag %VERSION% 到远程仓库...
git push origin %VERSION%

echo 代码和 tag %VERSION% 已成功推送到远程仓库!
endlocal
