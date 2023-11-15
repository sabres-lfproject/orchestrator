function orchestrator(name, image, cpu, memory, disk) {
    return {
        name: name,
        image: image,
        cpu: { cores: cpu, "passthru": true },
        memory: { capacity: GB(memory) },
        disks: [ { 'size': disk+"G", 'dev': 'vdb', 'bus': 'virtio' } ],
        mounts: [{ source: cwd+'/../..', point: "/tmp/orchestrator" }]
    };
}

function node(name, image, cpu, memory) {
    return {
        name: name,
        image: image,
        memory: { capacity: GB(memory) },
    };
}

topo = {
	name: "orchestrator"+Math.random().toString().substr(-6),
	nodes: [
		orchestrator("orchestrator","ubuntu-2204",8,32,64),
	],
    switches: [],
	links: []
}
