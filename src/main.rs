#![warn(clippy::str_to_string)]

mod commands;
mod config;
mod create;
mod role;

use crate::{config::CONFIG, role::random_color};

use poise::serenity_prelude as serenity;
use serenity::{
    all::{ActivityData, ClientBuilder, GatewayIntents, Permissions},
    builder::{CreateMessage, EditRole},
    FullEvent,
};
use std::{
    sync::{Arc, OnceLock},
    time::Duration,
};

type Error = Box<dyn std::error::Error + Send + Sync>;
type PoiseContext<'a> = poise::Context<'a, Data, Error>;

pub struct Data;

static FIRST_TIME: OnceLock<bool> = OnceLock::new();

async fn on_error(error: poise::FrameworkError<'_, Data, Error>) {
    match error {
        poise::FrameworkError::Setup { error, .. } => panic!("Failed to start bot: {:?}", error),
        poise::FrameworkError::Command { error, ctx, .. } => {
            println!("Error in command `{}`: {:?}", ctx.command().name, error,);
        }
        error => {
            poise::builtins::on_error(error).await.unwrap_or_else(|e| println!("Error while handling error: {}", e))
        }
    }
}

async fn event_handler(
    ctx: &serenity::Context,
    event: &serenity::FullEvent,
    _framework: poise::FrameworkContext<'_, Data, Error>,
    _: &Data,
) -> Result<(), Error> {
    match event {
        FullEvent::Ready { data_about_bot, .. } => {
            println!("Logged in as {}", data_about_bot.user.name);
            ctx.set_activity(Some(ActivityData::playing("Perfect Color!")));

            if FIRST_TIME.set(true).is_ok() {
                for guild in ctx.cache.guilds() {
                    let existing_roles: Vec<String> =
                        guild.roles(&ctx.http).await?.values().map(|role| role.name.clone()).collect();
                    for (name, &color) in &CONFIG.colors {
                        if !existing_roles.contains(name) {
                            let r = EditRole::new().name(name).colour(color).permissions(Permissions::empty());
                            match guild.create_role(&ctx.http, r).await {
                                Ok(_) => {}
                                Err(e) => {
                                    println!(
                                        "Error while creating colors on this server {} - {}, {e:?}",
                                        guild.get(),
                                        guild.name(&ctx.cache).unwrap()
                                    );
                                    break;
                                }
                            }
                        }
                    }
                }
            }
            println!("Bot is ready")
        }
        FullEvent::GuildMemberAddition { new_member } => {
            if let Err(e) = random_color(&ctx, &new_member).await {
                return Ok(println!("Error while member join, {}", e));
            }
            if !CONFIG.auto_kick.contains_key(&new_member.guild_id.to_string()) {
                return Ok(println!("Server not in kick list"));
            }

            tokio::time::sleep(Duration::from_secs(30 * 60)).await;

            match new_member.roles(&ctx.cache) {
                Some(roles) => {
                    if roles.iter().any(|role| CONFIG.colors.contains_key(&role.name)) {
                        return Ok(());
                    }
                }
                None => return Ok(eprintln!("Unable to retrieve roles from cache")),
            }

            if let Err(e) = new_member
                .user
                .dm(
                    &ctx.http,
                    CreateMessage::new().content(format!(
                        "You got kicked from the server.\nPlease read the welcome channel for more information\n{}",
                        CONFIG.invite_link
                    )),
                )
                .await
            {
                return Ok(println!("Unable to message user in private chat, {e:?}"));
            };

            if let Err(e) = new_member.kick_with_reason(&ctx.http, "User hasn't picked a role after 30 minutes").await {
                return Ok(println!("Unable to kick user after 30 minutes, {e:?}"));
            }
        }
        _ => {}
    }
    Ok(())
}

#[tokio::main]
async fn main() {
    env_logger::init();

    if CONFIG.bot_token == "YOUR_BOT_TOKEN" {
        return eprintln!("Your bot token is still the default, please change it in the config.yaml");
    }

    let intents = GatewayIntents::non_privileged() | GatewayIntents::MESSAGE_CONTENT | GatewayIntents::GUILD_MEMBERS;
    let options = poise::FrameworkOptions {
        commands: vec![commands::color(), commands::preview(), commands::help()],
        prefix_options: poise::PrefixFrameworkOptions {
            prefix: Some("<<".into()),
            edit_tracker: Some(Arc::from(poise::EditTracker::for_timespan(Duration::from_secs(3600)))),
            ..Default::default()
        },
        on_error: |error| Box::pin(on_error(error)),
        skip_checks_for_owners: false,
        event_handler: |ctx, event, framework, data| Box::pin(event_handler(ctx, event, framework, data)),
        ..Default::default()
    };

    let framework = poise::Framework::builder()
        .setup(move |ctx, _ready, framework| {
            Box::pin(async move {
                poise::builtins::register_globally(ctx, &framework.options().commands)
                    .await
                    .map_err(|e| e.into())
                    .map(|_| Data)
            })
        })
        .options(options)
        .build();

    let mut client = ClientBuilder::new(&CONFIG.bot_token, intents)
        .framework(framework)
        .await
        .expect("Couldn't build a bot client!");

    let shards = client.shard_manager.clone();
    tokio::spawn(async move {
        if let Ok(_) = tokio::signal::ctrl_c().await {
            shards.shutdown_all().await;
        }
    });

    if let Err(e) = client.start().await {
        println!("Bot crashed on {e:?}");
        client.shard_manager.shutdown_all().await
    }
}
