import { BrowserAutomation } from './browserAutomation';

async function example() {
    // Initialize browser automation
    const browser = new BrowserAutomation({
        headless: false,  // Set to true for production
        slowMo: 100      // Slow down operations for demonstration
    });

    try {
        // Start the browser
        await browser.initialize();

        // Navigate to a page
        await browser.navigate('https://example.com');

        // Wait for some element
        await browser.waitForSelector('h1');

        // Get text content
        const title = await browser.getText('h1');
        console.log('Page title:', title);

        // Take a screenshot
        await browser.screenshot('example.png');

    } catch (error) {
        console.error('Error:', error);
    } finally {
        // Always close the browser
        await browser.close();
    }
}

// Run the example
example().catch(console.error);
