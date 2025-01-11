#ifndef BLUETOOTH_WINDOWS_H
#define BLUETOOTH_WINDOWS_H

#include <winsock2.h>
#include <ws2bth.h>

#define  SUCCESS                     0

#define CXN_TRANSFER_DATA_LENGTH          8192
#define CXN_SUCCESS                       (0)
#define CXN_ERROR                         (-1)
#define CXN_DEFAULT_LISTEN_BACKLOG        4


SOCKET Listen(const GUID* serviceClassId, const char* serviceInstanceName, const char* comment);
SOCKET Accept(SOCKET severSocket);
int isInvalidSocket(SOCKET socket);
#endif // BLUETOOTH_WINDOWS_H

