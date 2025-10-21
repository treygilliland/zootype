// dinotype - TypeScript implementation of zootype typing test

const VERSION = "dev";

function main(): void {
  const args = Deno.args;

  if (args.length > 0 && (args[0] === "--version" || args[0] === "-v")) {
    console.log(`dinotype ${VERSION}`);
    return;
  }

  console.log("Hello from dinotype");
}

if (import.meta.main) {
  main();
}
