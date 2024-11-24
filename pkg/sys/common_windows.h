#ifndef COMMON_WINDOWS_H
#define COMMON_WINDOWS_H

#define ERROR_SOURCE_UNKNOWN          -1
#define ERROR_SOURCE_SUCCESS           0
#define ERROR_USER_LOGIN_FAILURE             10001
#define ERROR_USER_ACCOUNT_RESTRICTION       10002
#define ERROR_USER_WRONG_PASSWORD            10003
#define ERROR_USER_ACCOUNT_DISABLED          10004
#define ERROR_USER_INVALID_PARAMETER               10005

#define ERROR_SYSTEM_INSUFFICIENT_MEMORY             20001
#define ERROR_SYSTEM_CREDENTIAL_PROVIDER_ERROR       20002
#define ERROR_SYSTEM_PLUGIN_MANAGER_EXCEPTION        20003
#define ERROR_SYSTEM_INTERNAL_ERROR                  20004
#define ERROR_SYSTEM_PARAMETER_ERROR                 20005

#define ERROR_SYSTEM_BLUETOOTH_INIT_FAILURE          20007
#define ERROR_SYSTEM_BLUETOOTH_STOP_FAILURE          20008
#define ERROR_SYSTEM_SERVICE_ALREADY_RUNNING         20009




#define ERROR_UNKNOWN_LOGIN_FAILURE                 90001

#include <windows.h>
#include <stdlib.h>
#include <wchar.h>
#include <stdbool.h>
int TryLogin(wchar_t *username, wchar_t *pwd, wchar_t *domain);
int ConvertErrCode(DWORD r);
int PreCheckShutdownWindows();
int ShutdownWindows(UINT uFlags);
int PreCheckStandbyWindows();
int StandbyWindows();
int LockWindows();
int checkGrant();
int IsSessionLocked();
bool SetProcessPowerSavingMode(bool enable);
void set_process_priority(bool enable);

#endif // COMMON_WINDOWS_H
