using System;
using StackExchange.Redis;
using Neo.Ledger;
using Neo.VM;
using Neo.IO.Json;
using Neo.SmartContract;
using System.Collections.Generic;

using Snapshot = Neo.Persistence.Snapshot;

namespace Neo.Plugins
{
    public class NeoPubSub : Plugin, IPersistencePlugin
    {
        private readonly ConnectionMultiplexer connection;

        public NeoPubSub()
        {
            Console.WriteLine($"Connecting to PubSub server at {Settings.Default.RedisHost}:{Settings.Default.RedisPort}");
            this.connection = ConnectionMultiplexer.Connect($"{Settings.Default.RedisHost}:{Settings.Default.RedisPort}");
            if (this.connection == null) {
                Console.WriteLine("Connection failed!");
            } else {
                Console.WriteLine("Connected.");
            }
        }

        public override void Configure()
        {
            Settings.Load(GetConfiguration());
        }

        public void OnPersist(Snapshot snapshot, IReadOnlyList<Blockchain.ApplicationExecuted> applicationExecutedList)
        {
			JObject blockJson = snapshot.PersistingBlock.ToJson();
			blockJson["confirmations"] = 1;
            connection.GetSubscriber().Publish("blocks", blockJson.ToString());
            foreach (var appExec in applicationExecutedList)
            {
                var txid = appExec.Transaction.Hash.ToString();
                foreach (ApplicationExecutionResult p in appExec.ExecutionResults)
                {
                    if (!p.VMState.HasFlag(VMState.FAULT))
                    {
                        foreach (NotifyEventArgs q in p.Notifications)
                        {
                            string contract = q.ScriptHash.ToString();
                            string r = q.State.ToParameter().ToJson().ToString();
                            connection.GetSubscriber().Publish("events", $"{{\"contract\":\"{contract}\", \"txid\":\"{txid}\", \"call\":{r}}}");
                        }
                    }
                }
            }
        }

        public void OnCommit(Snapshot snapshot)
        {
        }

        public bool ShouldThrowExceptionFromCommit(Exception ex)
        {
            return false;
        }

    }
}
