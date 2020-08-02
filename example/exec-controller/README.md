# Exec Controller Example

1. Start daemon: `controllerbus daemon`
2. Apply in cli 1: `controllerbus client exec -f exec-controller.yaml`
3. Apply the same command in another CLI (concurrently).

Observe: 

 - Daemon output the text from the exampleField in the YAML
 - The second invocation returns RUNNING and does not print the same message
   again. This is due to directive de-duplication, it aliases the request to the
   already running controller.
 - If you ctrl+c the first invocation, it continues running due to the second.
 - Once both exec calls are canceled, the controller is unloaded.
 - Running step #2 and #3 again will start the controller and print the message.


