@echo off
chcp 65001
echo ========================================
echo JetGeo 项目构建和运行脚本
echo ========================================

REM 设置错误处理
setlocal enabledelayedexpansion

@REM C:\maven\apache-maven-3.9.11\bin
set PATH=C:\maven\apache-maven-3.9.11\bin;%PATH%

REM 检查 Maven 是否安装
where mvn >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo 错误: 未找到Maven，请确保Maven已安装并配置在PATH中
    echo.
    echo Maven安装指南:
    echo 1. 下载: https://maven.apache.org/download.cgi
    echo 2. 解压到 C:\apache-maven-3.9.x
    echo 3. 添加 C:\apache-maven-3.9.x\bin 到系统PATH
    echo 4. 重启命令行并运行 mvn -version 验证
    echo.
    echo 或使用包管理器:
    echo   choco install maven
    echo   scoop install maven
    pause
    exit /b 1
)

REM 检查 Java 是否安装
where java >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo 错误: 未找到Java，请确保JDK已安装并配置在PATH中
    pause
    exit /b 1
)

REM 切换到项目根目录
cd /d "%~dp0"

REM 显示环境信息
echo 当前环境信息:
java -version
echo.
mvn -version
echo.

echo ========================================
echo 步骤1: 清理之前的构建
echo ========================================
call mvn clean
if %ERRORLEVEL% neq 0 (
    echo 清理失败
    pause
    exit /b 1
)
echo 清理完成!
echo.

echo ========================================
echo 步骤2: 编译和打包项目
echo ========================================
call mvn compile package -DskipTests
if %ERRORLEVEL% neq 0 (
    echo 编译打包失败
    pause
    exit /b 1
)
echo 编译打包完成!
echo.

echo ========================================
echo 步骤3: 安装到本地仓库
echo ========================================
call mvn install -DskipTests
if %ERRORLEVEL% neq 0 (
    echo 安装到本地仓库失败
    pause
    exit /b 1
)
echo 安装到本地仓库完成!
echo.

echo ========================================
echo 步骤4: 检查地理数据文件
echo ========================================
if not exist "data\geodata" (
    echo 未找到解压的地理数据文件夹
    echo.
    if exist "data\geodata20250213.7z" (
        echo 发现 geodata20250213.7z 文件
        echo 请解压此文件到 data\geodata\ 目录 (推荐)
    ) else if exist "data\geodata.7z" (
        echo 发现 geodata.7z 文件  
        echo 请解压此文件到 data\geodata\ 目录
    ) else (
        echo 警告: 未找到地理数据文件!
        echo 请确保 data 目录下有 .7z 压缩文件
    )
    echo.
    echo 解压后的目录结构应该是: data\geodata\[省市区数据文件]
    echo 警告: 地理数据文件夹不存在，某些功能可能无法正常使用
) else (
    echo 地理数据文件夹已存在: data\geodata\
    echo.
)

echo ========================================
echo 步骤5: 运行测试示例
echo ========================================
echo 选择运行方式:
echo 1. 运行 JetGeo 核心示例 (JetGeoExample)
echo 2. 运行 JetGeo 单元测试 (JetGeoTests.initTest)
echo 3. 运行 Spring Boot 启动器测试 (JetGeoStarterApplicationTests)
echo 4. 跳过运行示例
echo.

set /p choice="请输入选择 (1-4): "

if "%choice%"=="1" (
    echo.
    echo 运行 JetGeo 核心示例...
    echo 注意: 需要先修改 JetGeoExample.java 中的数据路径
    cd jetgeo-core
    call mvn exec:java -Dexec.mainClass="com.ling5821.jetgeo.JetGeoExample"
    cd ..
) else if "%choice%"=="2" (
    echo.
    echo 运行 JetGeo 单元测试...
    cd jetgeo-core
    call mvn test -Dtest=JetGeoTests#initTest
    cd ..
) else if "%choice%"=="3" (
    echo.
    echo 运行 Spring Boot 启动器测试...
    cd jetgeo-spring-boot-starter
    call mvn test -Dtest=JetGeoStarterApplicationTests#testJetGeo
    cd ..
) else (
    echo 跳过运行示例
)

echo.
echo ========================================
echo 构建完成!
echo ========================================
echo.
echo 项目已成功构建并安装到本地Maven仓库
echo 版本: 1.2.2-SNAPSHOT
echo.
echo ============ 模块说明 ============
echo jetgeo-core: 
echo   - 核心地理信息查询库
echo   - 可独立使用
echo   - 支持省/市/区级别的坐标反向地理编码
echo.
echo jetgeo-spring-boot-starter:
echo   - Spring Boot自动配置模块  
echo   - 便于在Spring Boot项目中集成
echo.
echo ============ 使用方法 ============
echo.
echo 方法1: 在Spring Boot项目中使用
echo 在 pom.xml 中添加依赖:
echo   ^<dependency^>
echo     ^<groupId^>com.ling5821^</groupId^>
echo     ^<artifactId^>jetgeo-spring-boot-starter^</artifactId^>
echo     ^<version^>1.2.2-SNAPSHOT^</version^>
echo   ^</dependency^>
echo.
echo 在 application.yml 中配置:
echo   jetgeo:
echo     geoDataParentPath: %CD%\data\geodata
echo     level: [province, city, district]
echo.
echo 方法2: 直接使用核心库
echo 在 pom.xml 中添加依赖:
echo   ^<dependency^>
echo     ^<groupId^>com.ling5821^</groupId^>
echo     ^<artifactId^>jetgeo-core^</artifactId^>
echo     ^<version^>1.2.2-SNAPSHOT^</version^>
echo   ^</dependency^>
echo.
echo 方法3: 命令行测试
echo   cd jetgeo-core
echo   mvn exec:java -Dexec.mainClass="com.ling5821.jetgeo.JetGeoExample"
echo.
echo ============ 注意事项 ============
echo 1. 确保解压地理数据到 data\geodata\ 目录
echo 2. 示例代码中的数据路径可能需要根据实际情况调整
echo 3. 首次运行可能需要一些时间来加载地理数据
echo.
echo 构建脚本执行完成！