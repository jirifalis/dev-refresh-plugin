const RECONNECT_TIMEOUT = 1000; // 1s
const RECONNECT_LIMIT = 10;

const WS_SERVER_ADDR = 'ws://localhost:8888';
const RELOAD_PATTERN = 'http://localhost:8080/*';

let ws_plugin_connection;
let connection_limit;

resetConnectionLimit();
wsConnect();
initOnClickAction();

function initOnClickAction() {
    chrome.action.onClicked.addListener(function (tab) {
        resetConnectionLimit();
        wsConnect();
    });
}

function resetConnectionLimit() {
    connection_limit = RECONNECT_LIMIT;
}

function wsConnect() {
    if (connection_limit-- < 0) {
        return;
    }
    wsClose()
    ws_plugin_connection = new WebSocket(
        WS_SERVER_ADDR,
    );

    ws_plugin_connection.onopen = (event) => {
        console.log('Connection opened');
        iconStatusConnected();
    }
    ws_plugin_connection.onmessage = (event) => {
        console.log('Received message: ', event.data);
        reloadLocalhostTabs()
    };
    ws_plugin_connection.onclose = (event) => {
        console.log('Connection closed: ', event);
        iconStatusDisconnected();
        setTimeout(() => {
            wsConnect();
        }, RECONNECT_TIMEOUT);
    };
    ws_plugin_connection.onerror = (event) => {
        console.log('Connection error: ', event);
        iconStatusDisconnected();
        setTimeout(() => {
            wsConnect();
        }, RECONNECT_TIMEOUT);
    };
}

function wsClose() {
    if (ws_plugin_connection) {
        ws_plugin_connection.close();
        ws_plugin_connection = null;
        console.log("Connection terminated.");
    }
}


function reloadLocalhostTabs() {
    chrome.tabs.query({'url': RELOAD_PATTERN}, (tabs) => {
        tabs.forEach(tab => {
            console.log('Reloading tab: ', tab.id, tab.url);
            chrome.tabs.reload(tab.id, {bypassCache: true});
        })
    })
}

function iconStatusDisconnected() {
    iconStatusDraw("gray");
}

function iconStatusConnected() {
    iconStatusDraw("green");
}

function iconStatusDraw(color) {
    const canvas = new OffscreenCanvas(16, 16);
    const context = canvas.getContext('2d');

    context.arc(8, 8, 8, 0, 2 * Math.PI);
    context.fillStyle = color;
    context.fill();
    context.lineWidth = 1;
    context.strokeStyle = "black";
    context.stroke();
    const imageData = context.getImageData(0, 0, 16, 16);
    chrome.action.setIcon({imageData: imageData}, () => {
    });
}

