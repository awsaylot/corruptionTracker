import { chromium, Browser, Page } from 'playwright';

export interface BrowserOptions {
    headless?: boolean;
    slowMo?: number;
    userAgent?: string;
}

export class BrowserAutomation {
    private browser: Browser | null = null;
    private page: Page | null = null;

    constructor(private options: BrowserOptions = {}) {
        this.options = {
            headless: true,
            slowMo: 50,
            ...options
        };
    }

    async initialize(): Promise<void> {
        this.browser = await chromium.launch({
            headless: this.options.headless,
            slowMo: this.options.slowMo
        });

        const context = await this.browser.newContext({
            userAgent: this.options.userAgent
        });

        this.page = await context.newPage();
    }

    async navigate(url: string): Promise<void> {
        if (!this.page) {
            throw new Error('Browser not initialized. Call initialize() first.');
        }
        await this.page.goto(url);
    }

    async screenshot(path: string): Promise<void> {
        if (!this.page) {
            throw new Error('Browser not initialized. Call initialize() first.');
        }
        await this.page.screenshot({ path });
    }

    async click(selector: string): Promise<void> {
        if (!this.page) {
            throw new Error('Browser not initialized. Call initialize() first.');
        }
        await this.page.click(selector);
    }

    async type(selector: string, text: string): Promise<void> {
        if (!this.page) {
            throw new Error('Browser not initialized. Call initialize() first.');
        }
        await this.page.fill(selector, text);
    }

    async getText(selector: string): Promise<string | null> {
        if (!this.page) {
            throw new Error('Browser not initialized. Call initialize() first.');
        }
        return await this.page.textContent(selector);
    }

    async getHtml(selector: string): Promise<string | null> {
        if (!this.page) {
            throw new Error('Browser not initialized. Call initialize() first.');
        }
        return await this.page.innerHTML(selector);
    }

    async waitForSelector(selector: string): Promise<void> {
        if (!this.page) {
            throw new Error('Browser not initialized. Call initialize() first.');
        }
        await this.page.waitForSelector(selector);
    }

    async evaluate<T>(pageFunction: string | Function, ...args: any[]): Promise<T> {
        if (!this.page) {
            throw new Error('Browser not initialized. Call initialize() first.');
        }
        return await this.page.evaluate(pageFunction as any, ...args);
    }

    async close(): Promise<void> {
        if (this.browser) {
            await this.browser.close();
            this.browser = null;
            this.page = null;
        }
    }
}
