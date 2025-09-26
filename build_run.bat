@echo off
rem ============================================================================
rem  jetgeo helper script
rem  Usage: build_run.bat [command]
rem  Commands:
rem     build            = mvn clean install (skip tests by default; set RUN_TESTS=1 to run tests)
rem     build-full       = mvn clean install (always run tests)
rem     run-example      = compile and run example JetGeoExample in jetgeo-core (example is under test sources)
rem     run-server       = start spring boot server in jetgeo-server module
rem     test             = run all tests (mvn -DskipTests=false test)
rem     clean            = mvn clean
rem     deploy-snapshot  = deploy SNAPSHOT (credentials required in settings.xml server id=central)
rem     help             = show help
rem  Environment Variables:
rem     GEO_DATA_PATH    override geo data directory for example (example code currently hard-coded; modify source to use this)
rem     MAVEN_ARGS       append custom Maven args
rem     RUN_TESTS=1      enable tests in the 'build' command
rem ============================================================================

setlocal enabledelayedexpansion

set CMD=%1
if "%CMD%"=="" set CMD=build

rem Pre-check java & mvn
where java >nul 2>nul || (
    echo [ERROR] java not found, please install JDK and set PATH or JAVA_HOME
    exit /b 1
)
where mvn >nul 2>nul || (
    echo [ERROR] mvn not found, please install Maven and set PATH
    exit /b 1
)

set DEFAULT_SKIP=-DskipTests
if "%RUN_TESTS%"=="1" set DEFAULT_SKIP=

if /i "%CMD%"=="help" goto :help
if /i "%CMD%"=="build" goto :build
if /i "%CMD%"=="build-full" goto :build_full
if /i "%CMD%"=="run-example" goto :run_example
if /i "%CMD%"=="server" goto :run_server
if /i "%CMD%"=="test" goto :test
if /i "%CMD%"=="clean" goto :clean
if /i "%CMD%"=="deploy-snapshot" goto :deploy_snapshot

echo [ERROR] Unknown command: %CMD%
goto :help

:build
echo === Build (skip tests flag RUN_TESTS=%RUN_TESTS%) %DATE% %TIME% ===
mvn clean install %DEFAULT_SKIP% %MAVEN_ARGS%
if errorlevel 1 goto :fail
echo Build success
goto :eof

:build_full
echo === Build (with tests) ===
mvn clean install -DskipTests=false %MAVEN_ARGS%
if errorlevel 1 goto :fail
echo Build success (with tests)
goto :eof

:run_example
echo === Compile core module ===
mvn -q -pl jetgeo-core -am compile %MAVEN_ARGS%
if errorlevel 1 goto :fail
echo === Run JetGeoExample ===
rem Use test classpath because example class is under test sources
mvn -q -pl jetgeo-core -am exec:java -Dexec.classpathScope=test -Dexec.mainClass=com.ling5821.jetgeo.JetGeoExample -Dexec.cleanupDaemonThreads=false %MAVEN_ARGS%
if errorlevel 1 goto :fail
goto :eof

:run_server
echo === Package jetgeo-server ===
java -jar .\jetgeo-server\target\jetgeo-server-1.2.2-SNAPSHOT.jar --server.port=8080
goto :eof

:test
echo === Run tests ===
mvn -DskipTests=false test %MAVEN_ARGS%
if errorlevel 1 goto :fail
echo Tests finished
goto :eof

:clean
echo === Clean ===
mvn clean %MAVEN_ARGS%
if errorlevel 1 goto :fail
echo Clean done
goto :eof

:deploy_snapshot
echo === Deploy SNAPSHOT ===
echo NOTE: credentials for server id=central must be configured in %%USERPROFILE%%\.m2\settings.xml
mvn -DskipTests deploy %MAVEN_ARGS%
if errorlevel 1 goto :fail
echo SNAPSHOT deployed
goto :eof

:help
echo Usage: build_run.bat ^<command^>
echo Commands:
echo   build           build (skip tests by default; set RUN_TESTS=1 to run them)
echo   build-full      build and run tests
echo   run-example     run jetgeo-core example JetGeoExample
echo   run-server      run jetgeo-server Spring Boot HTTP service
echo   test            run tests only
echo   clean           clean
echo   deploy-snapshot deploy SNAPSHOT version
echo   help            show this help
echo Examples:
echo   RUN_TESTS=1 build_run.bat build
echo   build_run.bat run-example
goto :eof

:fail
echo =============================
echo [FAIL] Command failed (exit %errorlevel%)
echo =============================
exit /b %errorlevel%

endlocal
