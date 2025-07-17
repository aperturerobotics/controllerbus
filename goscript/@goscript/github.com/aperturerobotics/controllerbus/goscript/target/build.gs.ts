import * as $ from "@goscript/builtin/index.js";

import * as context from "@goscript/context/index.js"

export async function main(): Promise<void> {
	let ctx = context.Background()
	{
		let err = await run(ctx)
		if (err != null) {
			console.log("error:", err!.Error())
		}
	}
}

// run performs the filesystem operations previously done in the test
export async function run(ctx: context.Context): Promise<$.GoError> {
	let [subCtx, subCtxCancel] = context.WithCancel(ctx)
	queueMicrotask(() => {
		subCtxCancel!()
	})
	await $.chanRecv(subCtx!.Done())
	console.log("context test worked")

	return null
}

