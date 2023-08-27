package main

import (
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"

	"github.com/e-k-m/gomartini/internal/terrain"
	"github.com/e-k-m/gomartini/pkg/martini"
)

var addr = flag.String("addr", ":8080", "http service address")

var templ = template.Must(template.New("qr").Parse(templateStr))

type mesh struct {
	Vertices  []int
	Triangles []int
	Terrain   []float64
	Width     int
}

func main() {
	flag.Parse()
	http.Handle("/", http.HandlerFunc(m))
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func m(w http.ResponseWriter, req *http.Request) {

	terrain, err := terrain.Read("./fuji.png")

	if err != nil {
		panic(err)
	}

	m := martini.New(terrain.GridSize)
	tile := m.CreateTile(terrain.Terrain)
	vertices, triangles := tile.GetMesh(20)

	payload, err := json.Marshal(mesh{
		Vertices:  vertices,
		Triangles: triangles,
		Terrain:   terrain.Terrain,
		Width:     terrain.GridSize,
	})

	if err != nil {
		panic(err)
	}

	templ.Execute(w, string(payload))
}

const templateStr = `
<!DOCTYPE html>
<html>
    <head lang="en">
        <meta charset="utf-8">
        <title>Example Fuji</title>
        <style>
            body { margin: 0; }
        </style>
    </head>
    <body>
        
        <script async src="https://unpkg.com/es-module-shims@1.8.0/dist/es-module-shims.js"></script>

        <script type="importmap">
          {
            "imports": {
              "three": "https://unpkg.com/three@0.156.0/build/three.module.js",
              "three/addons/": "https://unpkg.com/three@0.156.0/examples/jsm/"
            }
          }
        </script>
       
        <script type="module">
        	function terrainGeometry(mesh) {
				const vertices = [];
				for (let i = 0; i < mesh.Vertices.length / 2; i++) {
					let x = mesh.Vertices[i * 2];
					let y = mesh.Vertices[i * 2 + 1];
					let z = mesh.Terrain[y * mesh.Width + x] / 50.0;
					vertices.push(x);
					vertices.push(z);
					vertices.push(y);
				}

				const geometry = new THREE.BufferGeometry();
				geometry.setIndex(new THREE.BufferAttribute(new Uint32Array(mesh.Triangles), 1));
				geometry.setAttribute(
					"position",
					new THREE.BufferAttribute(new Float32Array(vertices), 3),
				);

				geometry.attributes.position.needsUpdate = true;
				geometry.computeVertexNormals();
				geometry.computeBoundingBox();
				geometry.normalizeNormals();

				return geometry;
			}

			
            import * as THREE from 'three';
            import { OrbitControls } from 'three/addons/controls/OrbitControls.js';

            const mesh = JSON.parse({{.}})

            const renderer = new THREE.WebGLRenderer();
            renderer.setSize(window.innerWidth, window.innerHeight);
            document.body.appendChild(renderer.domElement);

            const scene = new THREE.Scene();
            scene.background = new THREE.Color("white");

            const camera = new THREE.PerspectiveCamera(55, window.innerWidth / window.innerHeight, 0.1, 500);
    		camera.position.set(80, 50, 150);

			const controls = new OrbitControls(camera, renderer.domElement);

            const geometry = terrainGeometry(mesh);

            const material = new THREE.MeshBasicMaterial({
                color: "black",
                wireframe: true,
            });
            
			const fuji = new THREE.Mesh(geometry, material);
            scene.add(fuji);

		    function animate() {
		      requestAnimationFrame(animate);
		      controls.update();
		      renderer.render(scene, camera);
		    }
		    animate();

        </script>       
    </body>
</html>
`
