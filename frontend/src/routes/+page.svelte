<script>
    import TerminalComp from '$lib/components/Terminal.svelte';
    let fileInput;
    let terminal;
    let zoomLevel = 1.0;
    let width = 180;
    let isStreaming = false;

    async function handleUpload() {
        const file = fileInput?.files[0];
        if (!file) return;

        isStreaming = true;
        terminal.clear();

        const formData = new FormData();
        formData.append("image", file);

        const isGif = file.type === 'image/gif';
        const endpoint = isGif ? 'stream' : 'convert';

        try {
            const response = await fetch(`http://localhost:8080/v1/${endpoint}?width=${width}&mode=ansi`, {
                method: 'POST',
                body: formData,
            });

            if (!response.ok) throw new Error('Falha no servidor');

            if (isGif) {
                const reader = response.body.pipeThrough(new TextDecoderStream()).getReader();
                while (true) {
                    const { value, done } = await reader.read();
                    if (done) break;
                    const cleanData = value.replace(/data: /g, '');
                    terminal.write(cleanData);
                }
            } else {
                const text = await response.text();
                terminal.write(text);
            }
        } catch (err) {
            terminal.write('\r\n\x1b[31mErro: ' + err.message + '\x1b[0m');
        } finally {
            isStreaming = false;
        }
    }
</script>

<div class="app-shell">
    <header class="controls-section">
        <div class="brand">RuneEngine <span>v1.0</span></div>
        
        <div class="actions">
            <div class="input-group">
                <input type="file" accept="image/*" bind:this={fileInput} />
            </div>
            
            <div class="input-group">
                <label for="w">Width:</label>
                <input type="number" id="w" bind:value={width} />
            </div>
            
            <div class="input-group">
                <label for="z">Zoom:</label>
                <input type="range" id="z" min="0.5" max="2" step="0.1" bind:value={zoomLevel} />
            </div>
            
            <button on:click={handleUpload} disabled={isStreaming}>
                {isStreaming ? '...' : 'Gerar'}
            </button>
        </div>
    </header>

    <section class="terminal-section">
        <TerminalComp bind:this={terminal} zoom={zoomLevel} />
    </section>
</div>

<style>
    :global(body, html) {
        margin: 0;
        padding: 0;
        height: 100vh;
        width: 100vw;
        background: #121212;
        color: white;
        overflow: hidden;
    }

    .app-shell {
        display: flex;
        flex-direction: column;
        height: 100vh;
        padding: 1rem;
        box-sizing: border-box;
    }

    .controls-section {
        flex: 0 0 auto;
        display: flex;
        justify-content: space-between;
        align-items: center;
        background: #1e1e1e;
        padding: 1rem;
        border-radius: 8px;
        margin-bottom: 1rem;
    }

    .actions { display: flex; gap: 1rem; align-items: center; }
    
    .input-group { display: flex; align-items: center; gap: 0.5rem; }
    
    input[type="number"] { width: 60px; background: #333; color: white; border: 1px solid #444; border-radius: 4px; }
    
    button { background: #ff3e00; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; }
    button:disabled { opacity: 0.5; }

    .terminal-section {
        flex: 1;
        min-height: 0;
        position: relative;
    }

    .brand { font-weight: bold; font-size: 1.2rem; }
    .brand span { color: #ff3e00; font-size: 0.8rem; }
</style>
