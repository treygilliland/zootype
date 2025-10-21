(* cameltype - OCaml implementation of zootype typing test *)

let version = "dev"

let () =
  if Array.length Sys.argv > 1 && (Sys.argv.(1) = "--version" || Sys.argv.(1) = "-v") then begin
    print_endline ("cameltype " ^ version);
    exit 0
  end;
  print_endline "Hello from cameltype"
