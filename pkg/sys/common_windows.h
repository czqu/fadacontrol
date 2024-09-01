#ifndef COMMON_WINDOWS_H
#define COMMON_WINDOWS_H

#include <windows.h>
#include <stdbool.h>

int PreCheckShutdownWindows();
int ShutdownWindows(UINT uFlags);
int PreCheckStandbyWindows();
int StandbyWindows();
int LockWindows();
int checkGrant();
int IsSessionLocked();
bool SetProcessPowerSavingMode(bool enable);
void set_process_priority(bool enable);
BOOL SetProcessInformation(
    HANDLE hProcess,
    PROCESS_INFORMATION_CLASS ProcessInformationClass,
    LPVOID ProcessInformation,
    DWORD ProcessInformationSize
);
#endif // COMMON_WINDOWS_H
