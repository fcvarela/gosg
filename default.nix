with import <nixpkgs> {};

stdenv.mkDerivation {
  name = "dev-environment";
  buildInputs = [pkgconfig xorg.libX11 xorg.libXi xorg.libXinerama xorg.libXcursor xorg.libXrandr xorg.libXxf86vm glfw3 bullet];
}

