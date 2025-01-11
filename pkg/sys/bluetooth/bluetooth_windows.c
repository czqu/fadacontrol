#include <winsock2.h>
#include <ws2bth.h>
#include "bluetooth_windwos.h"
#define  SUCCESS                     0

#define CXN_TRANSFER_DATA_LENGTH          8192
#define CXN_SUCCESS                       (0)
#define CXN_ERROR                         (-1)
#define CXN_DEFAULT_LISTEN_BACKLOG        4

SOCKET Listen(const GUID* serviceClassId, const char* serviceInstanceName, const char* comment)
{
    SOCKET severSocket = INVALID_SOCKET;
    WSADATA WSAData = {0};
    ULONG ulRetCode = CXN_SUCCESS;

    ulRetCode = WSAStartup(MAKEWORD(2, 2), &WSAData);
    if (ulRetCode == CXN_ERROR)
    {
        return INVALID_SOCKET;
    }
    int iAddrLen = sizeof(SOCKADDR_BTH);


    WSAQUERYSETW wsaQuerySet = {0};
    SOCKADDR_BTH SockAddrBthLocal = {0};
    LPCSADDR_INFO lpCSAddrInfo = NULL;

    lpCSAddrInfo = (LPCSADDR_INFO)HeapAlloc(GetProcessHeap(),
                                            HEAP_ZERO_MEMORY,
                                            sizeof(CSADDR_INFO));
    if (NULL == lpCSAddrInfo)
    {
        ulRetCode = CXN_ERROR;
        return INVALID_SOCKET;
    }


    if (CXN_SUCCESS == ulRetCode)
    {
        severSocket = socket(AF_BTH, SOCK_STREAM, BTHPROTO_RFCOMM);
        if (INVALID_SOCKET == severSocket)
        {
            ulRetCode = CXN_ERROR;
            return INVALID_SOCKET;
        }
    }

    if (CXN_SUCCESS == ulRetCode)
    {
        SockAddrBthLocal.addressFamily = AF_BTH;
        SockAddrBthLocal.port = BT_PORT_ANY;

        if (SOCKET_ERROR == bind(severSocket,
                                 (struct sockaddr*)&SockAddrBthLocal,
                                 sizeof(SOCKADDR_BTH)))
        {
            ulRetCode = CXN_ERROR;
            return INVALID_SOCKET;
        }
    }

    if (CXN_SUCCESS == ulRetCode)
    {
        ulRetCode = getsockname(severSocket,
                                (struct sockaddr*)&SockAddrBthLocal,
                                &iAddrLen);
        if (SOCKET_ERROR == ulRetCode)
        {
            ulRetCode = CXN_ERROR;
            return INVALID_SOCKET;
        }
    }
    if (CXN_SUCCESS == ulRetCode)
    {
        lpCSAddrInfo[0].LocalAddr.iSockaddrLength = sizeof(SOCKADDR_BTH);
        lpCSAddrInfo[0].LocalAddr.lpSockaddr = (LPSOCKADDR)&SockAddrBthLocal;
        lpCSAddrInfo[0].RemoteAddr.iSockaddrLength = sizeof(SOCKADDR_BTH);
        lpCSAddrInfo[0].RemoteAddr.lpSockaddr = (LPSOCKADDR)&SockAddrBthLocal;
        lpCSAddrInfo[0].iSocketType = SOCK_STREAM;
        lpCSAddrInfo[0].iProtocol = BTHPROTO_RFCOMM;

        ZeroMemory(&wsaQuerySet, sizeof(WSAQUERYSETW));
        wsaQuerySet.dwSize = sizeof(WSAQUERYSETW);
        wsaQuerySet.lpServiceClassId = (LPGUID)serviceClassId;
    }

    if (CXN_SUCCESS == ulRetCode)
    {
        size_t serviceInstanceNameLen = strlen(serviceInstanceName);
        size_t commentLen = strlen(comment);
        size_t convertedChars = 0;
        wchar_t* wServiceInstanceName = (wchar_t*)malloc((serviceInstanceNameLen + 1) * sizeof(wchar_t));
        wchar_t* wComment = (wchar_t*)malloc((commentLen + 1) * sizeof(wchar_t));

        if (!wServiceInstanceName || !wComment)
        {
            free(wServiceInstanceName);
            free(wComment);
            return INVALID_SOCKET;
        }


        mbstowcs_s(&convertedChars, wServiceInstanceName, serviceInstanceNameLen + 1, serviceInstanceName, _TRUNCATE);
        mbstowcs_s(&convertedChars, wComment, commentLen + 1, comment, _TRUNCATE);
        wsaQuerySet.lpszServiceInstanceName = wServiceInstanceName;
        wsaQuerySet.lpszComment = wComment;
        wsaQuerySet.dwNameSpace = NS_BTH;
        wsaQuerySet.dwNumberOfCsAddrs = 1;
        wsaQuerySet.lpcsaBuffer = lpCSAddrInfo;

        if (SOCKET_ERROR == WSASetServiceW(&wsaQuerySet, RNRSERVICE_REGISTER, 0))
        {
            ulRetCode = CXN_ERROR;
            return INVALID_SOCKET;
        }
    }
    if (SOCKET_ERROR == listen(severSocket, CXN_DEFAULT_LISTEN_BACKLOG))
    {
        return INVALID_SOCKET;
    }


    return severSocket;
}
SOCKET Accept(SOCKET severSocket)
{
    int clientSocket = accept(severSocket, NULL, NULL);

    return clientSocket;
}
int isInvalidSocket(SOCKET socket){
	return socket == INVALID_SOCKET;
}