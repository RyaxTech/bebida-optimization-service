with import <nixpkgs> { };

let
  customTex = texlive.combine {
    inherit (texlive)
      scheme-full
      textpos
      #fontspec
      marvosym
      todo;
  };
in
stdenv.mkDerivation rec {
  name = "atosStyleBeamer";
  buildInputs = with pkgs; [
    customTex
    rubber
    biber
    entr
    (python3.withPackages (ps: [ ps.pygments ]))
  ];
  src = ./.;
  slides = "./slides";
  FONTCONFIG_FILE = makeFontsConf { fontDirectories = [ corefonts ]; };
  buildPhase = "xelatex ${slides}.tex";
  installPhase = ''
    mkdir $out
    cp ${slides}.pdf $out
  '';
}
