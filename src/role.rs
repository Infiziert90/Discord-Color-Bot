use crate::{Error, PoiseContext};
use crate::config::CONFIG;

use std::collections::HashMap;
use std::hash::Hash;
use rand::thread_rng;
use serenity::http::CacheHttp;
use rand::seq::{IteratorRandom};
use serenity::all::{Member, Permissions};
use serenity::builder::EditRole;
use serenity::prelude::Context;

fn rand_hash<K: Eq + Hash, V>(hash: &HashMap<K, V>) -> &K {
    let key = hash.keys().choose(&mut thread_rng()).unwrap();

    key
}

pub async fn random_color(ctx: &Context, member: &Member) -> Result<(), Error> {
    let choice = rand_hash(&CONFIG.colors);

    let guild = match member.guild_id.to_guild_cached(&ctx.cache) {
        Some(g) => g.clone(),
        None => return Ok(())
    };

    let r = EditRole::new().name(choice.as_str()).colour(CONFIG.colors[choice]).permissions(Permissions::empty());
    let role_id = match guild.role_by_name(choice.as_str()) {
        Some(role) => role.id,
        None => {
            let r = guild.create_role(&ctx.http, r).await.unwrap();

            r.id
        }
    };

    let m = member.clone();
    match member.roles(&ctx.cache) {
        Some(roles) => {
            for role in roles {
                if CONFIG.colors.contains_key(role.name.as_str()) {
                    m.remove_role(ctx.http(), role.id).await?;
                    break;
                }
            }
        }
        None => {}
    };

    m.add_role(&ctx.http, role_id).await?;
    Ok(())
}

pub async fn process_color(ctx: PoiseContext<'_>, choice: String) -> Result<(), Error> {
    let guild = match ctx.guild() {
        Some(guild) => guild.clone(),
        None => {
            eprintln!("Can't find server ...");
            return Ok(())
        }
    };

    let r = EditRole::new().name(choice.as_str()).colour(CONFIG.colors[&choice]).permissions(Permissions::empty());
    let role_id = match guild.role_by_name(choice.as_str()) {
        Some(role) => role.id,
        None => {
            let r = guild.create_role(ctx.http(), r).await?;

            r.id
        }
    };

    let m = ctx.author_member().await.unwrap().to_mut().clone();
    match m.roles(ctx.cache()) {
        Some(roles) => {
            for role in roles {
                if CONFIG.colors.contains_key(role.name.as_str()) {
                    m.remove_role(ctx.http(), role.id).await?;
                    break;
                }
            }
        }
        None => {}
    };

    m.add_role(ctx.http(), role_id).await?;
    Ok(())
}