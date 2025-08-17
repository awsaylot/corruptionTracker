import { BrowserAutomation } from './browserAutomation';

async function main() {
    const args = process.argv.slice(2);
    const commandIndex = args.indexOf('--command');
    
    if (commandIndex === -1 || !args[commandIndex + 1]) {
        console.error('No command provided');
        process.exit(1);
    }

    let command;
    try {
        command = JSON.parse(args[commandIndex + 1]);
    } catch (error) {
        console.error('Invalid command JSON:', error);
        process.exit(1);
    }

    const browser = new BrowserAutomation({
        headless: true,  // Always run headless in production
        slowMo: 50
    });

    try {
        await browser.initialize();

        switch (command.type) {
            case 'navigate':
                await browser.navigate(command.options.url);
                break;

            case 'screenshot':
                await browser.screenshot(command.options.path);
                break;

            case 'click':
                await browser.click(command.options.selector);
                break;

            case 'type':
                await browser.type(command.options.selector, command.options.text);
                break;

            case 'getText':
                const text = await browser.getText(command.options.selector);
                console.log(JSON.stringify({ result: text }));
                break;

            default:
                throw new Error(`Unknown command type: ${command.type}`);
        }
    } catch (error) {
        console.error('Error executing command:', error);
        process.exit(1);
    } finally {
        await browser.close();
    }
}

main().catch(console.error);
