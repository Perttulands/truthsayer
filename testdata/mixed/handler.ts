import { Request, Response } from "express";

// Anti-pattern: as any
const config = {} as any;

export function handle(req: Request, res: Response) {
    const data = config.value;
    res.json(data);
}
