const redis = require("redis")
const WebSocket = require('ws');

const subscriber = redis.createClient()
const wss = new WebSocket.Server({ port: 8000 });

function broadcast(type, data){
	wss.clients.forEach(function each(client) {
		if (client.readyState === WebSocket.OPEN) {
			client.send(
				JSON.stringify({
					type,
					data
				})
			);
		}
	});
}

subscriber.on("message", function(channel, message) {
	broadcast(channel, JSON.parse(message));
});

subscriber.subscribe("events");
subscriber.subscribe("blocks");
