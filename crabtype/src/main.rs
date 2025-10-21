// crabtype - Rust implementation of zootype typing test

const VERSION: &str = "dev";

fn main() {
    let args: Vec<String> = std::env::args().collect();
    
    if args.len() > 1 && (args[1] == "--version" || args[1] == "-v") {
        println!("crabtype {}", VERSION);
        return;
    }
    
    println!("Hello from crabtype");
}

