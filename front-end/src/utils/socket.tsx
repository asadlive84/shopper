class WebSocketClient {
    private socket: WebSocket;
    private static instance: WebSocketClient;
  
    private constructor() {
      this.socket = new WebSocket('ws://localhost:8080/ws');
  
      this.socket.onopen = () => {
        console.log('WebSocket connected');
      };
  
      this.socket.onerror = (error) => {
        console.error('WebSocket error:', error);
      };
  
      this.socket.onclose = (event) => {
        console.log('WebSocket disconnected', event.code, event.reason);
      };
    }
  
    public static getInstance(): WebSocketClient {
      if (!WebSocketClient.instance) {
        WebSocketClient.instance = new WebSocketClient();
      }
      return WebSocketClient.instance;
    }
  
    public send(message: string) {
      if (this.socket.readyState === WebSocket.OPEN) {
        this.socket.send(message);
      } else {
        console.error('WebSocket is not open');
      }
    }
  
    public onMessage(callback: (event: MessageEvent) => void) {
      this.socket.onmessage = callback;
    }
  
    public close() {
      this.socket.close();
    }
  }
  
  const socket = WebSocketClient.getInstance();
  export default socket;