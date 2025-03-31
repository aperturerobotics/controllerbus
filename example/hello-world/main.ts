import { Context, Logger, background, ConsoleLogger } from '../../core/core.js'
import { newBus } from '../../bus/inmem/inmem.js'
import { newController as newDirectiveController } from '../../directive/controller/controller.js'
import {
  newController as newLoaderController,
  waitExecControllerRunning,
  newExecController,
} from '../../controller/loader/loader.js'
import { ToyControllerConfig } from './controller.pb'
import { newToyFactory } from './controller_factory'
import { ToyController } from './controller.js'
import './controller_config.js' // Import to add methods to ToyControllerConfig prototype

/**
 * Main function that runs the example.
 */
async function execToy(): Promise<void> {
  const ctx = background()
  const logger = new ConsoleLogger()

  // Create a directive controller
  const dc = newDirectiveController(ctx, logger)
  // Create a bus with the directive controller
  const b = newBus(dc)

  // Create a loader controller
  const [cl, err] = await newLoaderController(logger, b)
  if (err) {
    throw err
  }

  // Execute the loader controller
  const [releaseCl, addErr] = await b.addController(ctx, cl)
  if (addErr) {
    throw addErr
  }
  // releaseCl() would remove the loader controller

  logger.debug('loader controller attached')

  // Create a toy controller config
  const toyConfig = new ToyControllerConfig()
  toyConfig.name = 'world'

  // Issue directive to run the toy controller
  const loadToy = newExecController(newToyFactory(), toyConfig)

  try {
    const [ctrl, instance, valRef, execErr] = await waitExecControllerRunning(
      ctx,
      b,
      loadToy,
      null,
    )
    if (execErr) {
      throw execErr
    }

    const tc = ctrl as ToyController
    logger.debug('toy controller resolved')
    tc.sayHello()

    // Clean up
    valRef.release()
  } catch (e) {
    console.error(e)
  }
}

/**
 * Entry point for the example.
 */
function main(): void {
  execToy().catch((err) => {
    console.error(err)
    process.exit(1)
  })
}

// Run the example if this file is executed directly
if (require.main === module) {
  main()
}

export { main, execToy }
