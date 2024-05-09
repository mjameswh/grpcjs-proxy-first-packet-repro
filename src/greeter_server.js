const path = require("path");
const minimist = require("minimist");
const grpc = require("@grpc/grpc-js");
const protoLoader = require("@grpc/proto-loader");

const packageDefinition = protoLoader.loadSync(
  path.resolve(__dirname, "../protos/helloworld.proto")
);
const hello_proto = grpc.loadPackageDefinition(packageDefinition).helloworld;

function main() {
  var argv = minimist(process.argv.slice(2));
  const port = argv.port ?? 6233;
  if (argv._.length > 0) {
    console.log("Usage: node greeter_server.js [--port=PORT]");
    process.exit(1);
  }

  const server = new grpc.Server();
  server.addService(hello_proto.GreeterService.service, {
    sayHello: (call, callback) => {
      callback(null, { message: `Hello, ${call.request.name}` });
    },
  });
  server.bindAsync(`localhost:${port}`, grpc.ServerCredentials.createInsecure(), (err, port) => {
    if (err != null) {
      return console.error(err);
    }
    console.log(`gRPC listening on ${port}`);
  });
}

main();
