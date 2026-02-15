<script>
    import { onMount, onDestroy } from 'svelte';
    import '@xterm/xterm/css/xterm.css';

    let terminalElement;
    let term;
    let fitAddon;
    let resizeObserver;
    
    export let zoom = 1.0;
    const baseFontSize = 12;

    $: if (term && fitAddon && zoom) {
        term.options.fontSize = baseFontSize * zoom;
        setTimeout(() => fitAddon.fit(), 20);
    }

    export function write(data) { if (term) term.write(data); }
    export function clear() {
        if (term) {
            term.clear();
            term.write('\x1b[H');
        }
    }

    onMount(async () => {
        const { Terminal } = await import('@xterm/xterm');
        const { FitAddon } = await import('@xterm/addon-fit');
        
        term = new Terminal({
            theme: { background: '#000000' },
            convertEol: true,
            disableStdin: true,
            scrollback: 5000,
        });

        fitAddon = new FitAddon();
        term.loadAddon(fitAddon);
        term.open(terminalElement);

        resizeObserver = new ResizeObserver(() => {
            if (fitAddon) fitAddon.fit();
        });
        resizeObserver.observe(terminalElement);

        setTimeout(() => fitAddon.fit(), 50);
    });

    onDestroy(() => {
        if (resizeObserver) resizeObserver.disconnect();
    });
</script>

<div class="terminal-wrapper" bind:this={terminalElement}></div>

<style>
    .terminal-wrapper { 
        width: 100%;
        height: 100%;
        background: #000;
        border-radius: 4px;
    }

    :global(.xterm-viewport) {
        border-radius: 4px;
    }
</style>
