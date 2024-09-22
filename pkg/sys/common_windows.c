#include <windows.h>
#include <processthreadsapi.h>
#include <stdbool.h>
#include "common_windows.h"
int PreCheckShutdownWindows(){
// Get a token for this process.
    HANDLE hToken;
    TOKEN_PRIVILEGES tkp;

    if (!OpenProcessToken(GetCurrentProcess(),
        TOKEN_ADJUST_PRIVILEGES | TOKEN_QUERY, &hToken)) {
        return GetLastError(); // OpenProcessToken failed
    }

    // Get the LUID for the shutdown privilege.
    LookupPrivilegeValue(NULL, SE_SHUTDOWN_NAME,
        &tkp.Privileges[0].Luid);

    tkp.PrivilegeCount = 1;  // one privilege to set
    tkp.Privileges[0].Attributes = SE_PRIVILEGE_ENABLED;

    // Get the shutdown privilege for this process.
    AdjustTokenPrivileges(hToken, FALSE, &tkp, 0,
        (PTOKEN_PRIVILEGES)NULL, 0);

    if (GetLastError() != ERROR_SUCCESS) {
        return GetLastError(); // AdjustTokenPrivileges failed
    }
    return 0;
}
int ShutdownWindows( UINT  uFlags) {
	int ret=PreCheckShutdownWindows();
	if(ret!=0){
		return ret;
	}

    // Shut down the system and force all applications to close.
//If the function succeeds, the return value is nonzero. Because the function executes asynchronously,
//a nonzero return value indicates that the shutdown has been initiated. It does not indicate whether the shutdown will succeed.
//It is possible that the system, the user, or another application will abort the shutdown.
    if (!ExitWindowsEx(uFlags,
        SHTDN_REASON_MAJOR_OTHER  |
        SHTDN_REASON_MINOR_OTHER )) {
        return GetLastError(); // ExitWindowsEx failed
    }

    return 0; // Success
}
int PreCheckStandbyWindows() {
// Get a token for this process.
    HANDLE hToken;
    TOKEN_PRIVILEGES tkp;

    if (!OpenProcessToken(GetCurrentProcess(),
        TOKEN_ADJUST_PRIVILEGES | TOKEN_QUERY, &hToken)) {
        return GetLastError(); // OpenProcessToken failed
    }

    // Get the LUID for the hibernate privilege.
    LookupPrivilegeValue(NULL, SE_SHUTDOWN_NAME,
        &tkp.Privileges[0].Luid);

    tkp.PrivilegeCount = 1;  // one privilege to set
    tkp.Privileges[0].Attributes = SE_PRIVILEGE_ENABLED;

    // Get the hibernate privilege for this process.
    AdjustTokenPrivileges(hToken, FALSE, &tkp, 0,
        (PTOKEN_PRIVILEGES)NULL, 0);

    if (GetLastError() != ERROR_SUCCESS) {
        return GetLastError(); // AdjustTokenPrivileges failed
    }
    return 0; // Success
}
int StandbyWindows() {
	int ret=PreCheckStandbyWindows();
	if(ret!=0){
		return ret;
	}
    // Put the system into standby. If power has been suspended and subsequently restored, the return value is nonzero.
    if (!SetSystemPowerState(TRUE, TRUE)) {
        return GetLastError(); // SetSystemPowerState failed
    }

    return 0; // Success
}
int LockWindows() {

    if (LockWorkStation()) {

        return 0;
    } else {

        return GetLastError();
    }
}
int checkGrant(){
    HANDLE hToken;
    TOKEN_PRIVILEGES tkp;

    if (!OpenProcessToken(GetCurrentProcess(),
        TOKEN_ADJUST_PRIVILEGES | TOKEN_QUERY, &hToken)) {
        return GetLastError(); // OpenProcessToken failed
    }

    // Get the LUID for the hibernate privilege.
    LookupPrivilegeValue(NULL, SE_SHUTDOWN_NAME,
        &tkp.Privileges[0].Luid);

    tkp.PrivilegeCount = 1;  // one privilege to set
    tkp.Privileges[0].Attributes = SE_PRIVILEGE_ENABLED;

    // Get the hibernate privilege for this process.
    AdjustTokenPrivileges(hToken, FALSE, &tkp, 0,
        (PTOKEN_PRIVILEGES)NULL, 0);

    if (GetLastError() != ERROR_SUCCESS) {
        return GetLastError(); // AdjustTokenPrivileges failed
    }

	return 0;
}



int IsSessionLocked()
{
return GetSystemMetrics(SM_REMOTESESSION) ;
}

#if defined(__MINGW32__)
    bool SetProcessPowerSavingMode(bool enable) {
        return true;
    }
#else



bool SetProcessPowerSavingMode(bool enable) {
    //
    HANDLE hProcess = GetCurrentProcess();

    //
    PROCESS_POWER_THROTTLING_STATE PowerThrottling = {0};
    PowerThrottling.Version = PROCESS_POWER_THROTTLING_CURRENT_VERSION;
    PowerThrottling.ControlMask = PROCESS_POWER_THROTTLING_EXECUTION_SPEED;

    if (enable) {
    set_process_priority(true);
      PowerThrottling.ControlMask = PROCESS_POWER_THROTTLING_EXECUTION_SPEED;
        PowerThrottling.StateMask = PROCESS_POWER_THROTTLING_EXECUTION_SPEED;
    } else {
    set_process_priority(false);
        PowerThrottling.StateMask = 0;
        PowerThrottling.ControlMask = 0;
    }


    if (SetProcessInformation(hProcess, ProcessPowerThrottling, &PowerThrottling, sizeof(PowerThrottling))) {
        return true;
    } else {
        return false;
    }
}
#endif
void set_process_priority(bool enable)
{
    if (enable) {
        SetPriorityClass(GetCurrentProcess(), IDLE_PRIORITY_CLASS);
    }else{
        SetPriorityClass(GetCurrentProcess(), NORMAL_PRIORITY_CLASS);
    }


}