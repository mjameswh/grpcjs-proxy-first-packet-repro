const path = require("path");
const minimist = require("minimist");
const grpc = require("@grpc/grpc-js");
const protoLoader = require("@grpc/proto-loader");

const packageDefinition = protoLoader.loadSync(
  path.resolve(__dirname, "../protos/helloworld.proto")
);
const hello_proto = grpc.loadPackageDefinition(packageDefinition).helloworld;

async function callService(target) {
  const client = new hello_proto.GreeterService(target, grpc.credentials.createInsecure());
  try {
    const response = await new Promise((resolve, reject) =>
      client.sayHello({ name: "world" }, (err, response) => {
        if (err) reject(err);
        else resolve(response);
      })
    );
    // console.log("Greeting:", response.message);
    return false;
  } catch (e) {
    console.log(e);
    return true;
  } finally {
    client.close();
  }
}

async function main() {
  var argv = minimist(process.argv.slice(2), { string: ["target"] });
  const target = argv.target ?? "127.0.0.1:6233";
  const repeat = argv.repeat ?? 100;

  for (let i = 0; i < repeat; i++) {
    if (await callService(target)) {
      console.log(`Failed at iteration ${i + 1}`);
      return;
    }
    // This is just to make it easier to read GRPC traces
    if (process.env.GRPC_TRACE || process.env.GRPC_VERBOSITY) {
      console.log(`=============================`);
    }
  }

  console.log("Completed without error");
}

main().catch(console.error);
